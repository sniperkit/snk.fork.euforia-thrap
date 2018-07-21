package core

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
	"github.com/pkg/errors"

	"github.com/euforia/pseudo"
	"github.com/euforia/pseudo/scope"
	"github.com/euforia/thrap/consts"
	"github.com/euforia/thrap/crt"
	"github.com/euforia/thrap/orchestrator"

	"github.com/euforia/thrap/asm"
	"github.com/euforia/thrap/config"
	"github.com/euforia/thrap/packs"
	"github.com/euforia/thrap/registry"
	"github.com/euforia/thrap/store"
	"github.com/euforia/thrap/thrapb"
	"github.com/euforia/thrap/utils"
	"github.com/euforia/thrap/vcs"
)

var (
	// ErrStackAlreadyRegistered is used when a stack is already registered
	ErrStackAlreadyRegistered = errors.New("stack already registered")
	errComponentNotBuildable  = errors.New("component not buildable")
)

// Stack provides various stack based operations
type Stack struct {
	// builder is docker runtime
	crt *crt.Docker
	// config to use for this instance
	conf *config.ThrapConfig
	// there can only be one vcs provider
	vcs vcs.VCS
	// available registries
	regs map[string]registry.Registry
	// runtime orchestrator
	orch orchestrator.Orchestrator
	// packs
	packs *packs.Packs
	// stack store
	sst StackStorage

	log *log.Logger
}

// Assembler returns a new assembler for the stack
func (st *Stack) Assembler(cwd string, stack *thrapb.Stack) (*asm.StackAsm, error) {
	scopeVars := st.conf.VCS[st.vcs.ID()].ScopeVars("vcs.")
	return asm.NewStackAsm(stack, cwd, st.vcs, nil, scopeVars, st.packs)
}

// Register registers a new stack. It returns an error if the stack is
// already registered or fails to register
func (st *Stack) Register(stack *thrapb.Stack) (*thrapb.Stack, []*ActionReport, error) {
	errs := stack.Validate()
	if len(errs) > 0 {
		return nil, nil, utils.FlattenErrors(errs)
	}

	stack, err := st.sst.Create(stack)
	if err != nil {
		if err == store.ErrStackExists {
			return nil, nil, ErrStackAlreadyRegistered
		}
		return nil, nil, err
	}

	reports := st.ensureStackResources(stack)

	// Temp
	for _, r := range reports {
		fmt.Printf("%v '%v'\n", r.Action, r.Error)
	}

	return stack, reports, err
}

// Validate validates the stack manifest
func (st *Stack) Validate(stack *thrapb.Stack) error {
	// stack.Version = vcs.GetRepoVersion(ctxDir).String()
	errs := stack.Validate()
	if len(errs) > 0 {
		return utils.FlattenErrors(errs)
	}
	return nil
}

// Commit updates a stack definition
func (st *Stack) Commit(stack *thrapb.Stack) (*thrapb.Stack, error) {
	errs := stack.Validate()
	if len(errs) > 0 {
		return nil, utils.FlattenErrors(errs)
	}

	return st.sst.Update(stack)
}

func (st *Stack) populateFromImageConf(stack *thrapb.Stack) {
	confs := st.getContainerConfigs(stack)
	st.populatePorts(stack, confs)
	st.populateVolumes(stack, confs)
}

// populatePorts populates ports into the stack from the container images for ports
// that have not been defined in the stack but are in the image config
func (st *Stack) populatePorts(stack *thrapb.Stack, contConfs map[string]*container.Config) {
	for id, cfg := range contConfs {
		comp := stack.Components[id]
		if comp.Ports == nil {
			comp.Ports = make(map[string]int32, len(cfg.ExposedPorts))
		}

		if len(cfg.ExposedPorts) == 1 {
			for k := range cfg.ExposedPorts {
				if !comp.HasPort(int32(k.Int())) {
					comp.Ports["default"] = int32(k.Int())
				}
				break
			}
		} else {
			for k := range cfg.ExposedPorts {
				if !comp.HasPort(int32(k.Int())) {
					// HCL does not allow numbers as keys
					comp.Ports["port"+k.Port()] = int32(k.Int())
				}
			}
		}
	}
}

func (st *Stack) populateVolumes(stack *thrapb.Stack, contConfs map[string]*container.Config) {
	for id, cfg := range contConfs {
		comp := stack.Components[id]
		vols := make([]*thrapb.Volume, 0, len(cfg.Volumes))

		for k := range cfg.Volumes {
			if !comp.HasVolumeTarget(k) {
				vols = append(vols, &thrapb.Volume{Target: k})
			}
		}
		comp.Volumes = append(comp.Volumes, vols...)
	}
}

func (st *Stack) getContainerConfigs(stack *thrapb.Stack) map[string]*container.Config {
	out := make(map[string]*container.Config, len(stack.Components))
	for _, comp := range stack.Components {
		// Ensure we have the image locally
		err := st.crt.ImagePull(context.Background(), comp.Name+":"+comp.Version)
		if err != nil {
			continue
		}
		// Get image config
		ic, err := st.crt.ImageConfig(comp.Name + ":" + comp.Version)
		if err != nil {
			continue
		}

		out[comp.ID] = ic
	}
	return out
}

// Init initializes a basic stack with the configuration and options provided. This should only be
// used in the local cli case as the config is merged with the global.
func (st *Stack) Init(stconf *asm.BasicStackConfig, opt ConfigureOptions) (*thrapb.Stack, error) {

	_, err := ConfigureLocal(st.conf, opt)
	if err != nil {
		return nil, err
	}

	repo := opt.VCS.Repo
	vcsp, gitRepo, err := vcs.SetupLocalGitRepo(repo.Name, repo.Owner, opt.DataDir, opt.VCS.Addr)
	if err != nil {
		return nil, err
	}

	stack, err := asm.NewBasicStack(stconf, st.packs)
	if err != nil {
		return nil, err
	}
	if errs := stack.Validate(); len(errs) > 0 {
		return nil, utils.FlattenErrors(errs)
	}

	st.populateFromImageConf(stack)

	scopeVars := st.conf.VCS[st.vcs.ID()].ScopeVars("vcs.")
	stasm, err := asm.NewStackAsm(stack, opt.DataDir, vcsp, gitRepo, scopeVars, st.packs)
	if err != nil {
		return stack, err
	}

	err = stasm.AssembleMaterialize()
	if err == nil {
		err = stasm.WriteManifest()
	}

	return stack, err
}

func (st *Stack) scopeVars(stack *thrapb.Stack) scope.Variables {
	svars := stack.ScopeVars()
	for k, v := range stack.Components {

		ipvar := ast.Variable{
			Type:  ast.TypeString,
			Value: k + "." + stack.ID,
		}

		// Set container ip var
		svars[v.ScopeVarName(consts.CompVarPrefixKey+".", "container.ip")] = ipvar
		// Set container.addr var per port label
		for pl, p := range v.Ports {
			svars[v.ScopeVarName(consts.CompVarPrefixKey+".", "container.addr."+pl)] = ast.Variable{
				Type:  ast.TypeString,
				Value: fmt.Sprintf("%s:%d", ipvar.Value, p),
			}
		}

	}

	return svars
}

// startServices starts services needed to perform the build that themselves do not need
// to be built
func (st *Stack) startServices(ctx context.Context, stack *thrapb.Stack, scopeVars scope.Variables) error {
	var (
		err error
		//opt = crt.RequestOptions{Output: os.Stdout}
	)

	fmt.Printf("\nServices:\n\n")

	for _, comp := range stack.Components {
		if comp.IsBuildable() {
			continue
		}

		// eval hcl/hil
		if err = st.evalComponent(comp, scopeVars); err != nil {
			break
		}

		// Pull image if we do not locally have it
		imageID := comp.Name + ":" + comp.Version
		if !st.crt.HaveImage(ctx, imageID) {
			err = st.crt.ImagePull(ctx, imageID)
			if err != nil {
				break
			}
		}

		if err = st.startContainer(ctx, stack.ID, comp); err != nil {
			break
		}

		fmt.Println(comp.ID)

	}

	return err
}

func (st *Stack) startContainer(ctx context.Context, sid string, comp *thrapb.Component) error {
	cfg := thrapb.NewContainer(sid, comp.ID)

	if comp.IsBuildable() {
		cfg.Container.Image = filepath.Join(sid, comp.Name)
	} else {
		cfg.Container.Image = comp.Name
	}

	// Add image version if present
	if len(comp.Version) > 0 {
		cfg.Container.Image += ":" + comp.Version
	}

	if comp.HasEnvVars() {
		cfg.Container.Env = make([]string, 0, len(comp.Env.Vars))
		for k, v := range comp.Env.Vars {
			cfg.Container.Env = append(cfg.Container.Env, k+"="+v)
		}
	}

	// Publish all ports for a head component.
	// TODO: May need to map this to user defined host ports
	if comp.Head {
		cfg.Host.PublishAllPorts = true
	}

	// Non-blocking
	warnings, err := st.crt.Run(ctx, cfg)
	if err != nil {
		return err
	}

	if len(warnings) > 0 {
		for _, w := range warnings {
			fmt.Printf("%s: %s\n", cfg.Name, w)
		}
	}

	return nil
}

// Log writes the log for a running component to the writers
func (st *Stack) Log(ctx context.Context, id string, stdout, stderr io.Writer) error {
	return st.crt.Logs(ctx, id, stdout, stderr)
}

// Logs writes all running component logs for the stack
func (st *Stack) Logs(ctx context.Context, stack *thrapb.Stack, stdout, stderr io.Writer) error {
	var err error
	for _, comp := range stack.Components {
		er := st.crt.Logs(ctx, comp.ID+"."+stack.ID, stdout, stderr)
		if er != nil {
			err = er
		}
	}

	return err
}

// Status returns a CompStatus slice containing the status of each component
// in the stack
func (st *Stack) Status(ctx context.Context, stack *thrapb.Stack) []*CompStatus {
	out := make([]*CompStatus, 0, len(stack.Components))
	for _, comp := range stack.Components {
		id := comp.ID + "." + stack.ID
		ss := st.getCompStatus(ctx, id)
		ss.ID = comp.ID

		out = append(out, ss)
	}

	return out
}

func (st *Stack) getCompStatus(ctx context.Context, id string) *CompStatus {

	ss := &CompStatus{}
	ss.Details, ss.Error = st.crt.Inspect(ctx, id)

	if ss.Error == nil {
		if ss.Details.State.Status == "exited" {
			s := ss.Details.State
			ss.Error = fmt.Errorf("code=%d", s.ExitCode)
		}

	} else {
		ss.Details = types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{Status: "failed"},
			},
			Config: &container.Config{},
		}
	}

	return ss
}

// Images returns all known images for the stack
func (st *Stack) Images(stack *thrapb.Stack) []*CompImage {
	images := make([]*CompImage, 0, len(stack.Components))

	ctx := context.Background()
	conts, err := st.crt.ListImagesWithLabel(ctx, "stack="+stack.ID)
	if err != nil {
		return images
	}

	for _, c := range conts {
		ci := NewCompImage(c.ID, c.RepoTags)
		ci.Labels = c.Labels
		ci.Created = time.Unix(c.Created, 0)
		ci.Size = c.Size
		images = append(images, ci)
	}

	return images
}

// Get returns a stack from the store by id
func (st *Stack) Get(id string) (*thrapb.Stack, error) {
	return st.sst.Get(id)
}

// Iter iterates over each stack definition in the store.
func (st *Stack) Iter(prefix string, f func(*thrapb.Stack) error) error {
	return st.sst.Iter(prefix, f)
}

// Build starts all require services, then starts all the builds
func (st *Stack) Build(ctx context.Context, stack *thrapb.Stack) error {
	if errs := stack.Validate(); len(errs) > 0 {
		return utils.FlattenErrors(errs)
	}

	err := st.crt.CreateNetwork(ctx, stack.ID)
	if err != nil {
		return err
	}

	scopeVars := st.scopeVars(stack)
	fmt.Printf("\nScope:\n\n")
	fmt.Println(strings.Join(scopeVars.Names(), "\n"))
	fmt.Println()

	defer st.Destroy(ctx, stack)

	// Start containers needed for build
	err = st.startServices(ctx, stack, scopeVars)
	if err != nil {
		return err
	}

	// Start non-head builds
	err = st.startBuilds(ctx, stack, scopeVars, false)
	if err != nil {
		return err
	}

	// Start head builds
	err = st.startBuilds(ctx, stack, scopeVars, true)

	return err
}

// Destroy removes call components of the stack from the container runtime
func (st *Stack) Destroy(ctx context.Context, stack *thrapb.Stack) []*ActionReport {
	ar := make([]*ActionReport, 0, len(stack.Components))

	for _, c := range stack.Components {
		r := &ActionReport{Action: NewAction("destroy", "comp", c.ID)}
		r.Error = st.crt.Remove(ctx, c.ID+"."+stack.ID)
		ar = append(ar, r)
	}
	return ar
}

// Stop shutsdown any running containers in the stack.
func (st *Stack) Stop(ctx context.Context, stack *thrapb.Stack) []*ActionReport {
	ar := make([]*ActionReport, 0, len(stack.Components))

	for _, c := range stack.Components {
		r := &ActionReport{Action: NewAction("stop", "comp", c.ID)}
		r.Error = st.crt.Stop(ctx, c.ID+"."+stack.ID)
		ar = append(ar, r)
	}
	return ar
}

// Deploy deploys all components of the stack.
func (st *Stack) Deploy(stack *thrapb.Stack) error {
	if errs := stack.Validate(); len(errs) > 0 {
		return utils.FlattenErrors(errs)
	}

	var (
		ctx = context.Background()
		err = st.crt.CreateNetwork(ctx, stack.ID)
	)

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			st.Destroy(ctx, stack)
		}
	}()

	svars := st.scopeVars(stack)

	// Deploy services like db's etc
	err = st.startServices(ctx, stack, svars)
	if err != nil {
		return err
	}

	// Deploy non-head containers
	for _, comp := range stack.Components {
		if !comp.IsBuildable() {
			continue
		}

		if comp.Head {
			continue
		}

		if err = st.evalComponent(comp, svars); err != nil {
			return err
		}

		err = st.startContainer(ctx, stack.ID, comp)
		if err != nil {
			return err
		}

		fmt.Println(comp.ID)
	}

	// Start head containers
	for _, comp := range stack.Components {
		if !comp.IsBuildable() {
			continue
		}
		if !comp.Head {
			continue
		}

		if err = st.evalComponent(comp, svars); err != nil {
			break
		}

		err = st.startContainer(ctx, stack.ID, comp)
		if err != nil {
			break
		}

		fmt.Println(comp.ID)
	}

	return err
}

func (st *Stack) startBuilds(ctx context.Context, stack *thrapb.Stack, scopeVars scope.Variables, head bool) error {
	var err error

	// Start build containers after
	for _, comp := range stack.Components {
		if !comp.IsBuildable() {
			continue
		}

		// Build based on whether head was requested
		if comp.Head != head {
			continue
		}

		// eval hcl/hil
		if err = st.evalComponent(comp, scopeVars); err != nil {
			break
		}

		// err = st.buildStages(ctx, stack.ID, comp)
		_, err = st.doBuild(ctx, stack.ID, comp)
		if err != nil {
			break
		}

		// Start container from image that was just built, if this component
		// is not the head
		if !comp.Head {
			err = st.startContainer(ctx, stack.ID, comp)
			if err != nil {
				break
			}
		}
	}

	return err
}

// getBuildImageTags returns tags that should be applied to a given image build
func (st *Stack) getBuildImageTags(stackID string, comp *thrapb.Component) []string {
	base := filepath.Join(stackID, comp.ID)
	out := []string{base}
	if len(comp.Version) > 0 {
		out = append(out, base+":"+comp.Version)
	}

	rconf := st.conf.GetDefaultRegistry()
	if rconf != nil && len(rconf.Addr) > 0 {
		rbase := filepath.Join(rconf.Addr, base)
		out = append(out, rbase)
		if len(comp.Version) > 0 {
			out = append(out, rbase+":"+comp.Version)
		}
	}
	return out
}

func (st *Stack) makeBuildRequest(sid string, comp *thrapb.Component) *crt.BuildRequest {
	req := &crt.BuildRequest{
		Output:     os.Stdout,
		ContextDir: comp.Build.Context,
		BuildOpts: &types.ImageBuildOptions{
			Tags: st.getBuildImageTags(sid, comp),
			// ID to use in order to cancel the build
			BuildID:     comp.ID,
			Dockerfile:  comp.Build.Dockerfile,
			NetworkMode: sid,
			// Add labels to query later
			Labels: map[string]string{
				"stack":     sid,
				"component": comp.ID,
			},
		},
	}

	if comp.HasEnvVars() {
		args := make(map[string]*string, len(comp.Env.Vars))

		fmt.Printf("\nBuild arguments:\n\n")
		for k := range comp.Env.Vars {
			fmt.Println(k)

			v := comp.Env.Vars[k]
			args[k] = &v
		}
		fmt.Println()

		req.BuildOpts.BuildArgs = args
	}

	return req
}

func (st *Stack) doBuild(ctx context.Context, sid string, comp *thrapb.Component) (map[string]string, error) {
	req := st.makeBuildRequest(sid, comp)

	fmt.Printf("Building %s:\n\n", comp.ID)

	// Blocking
	err := st.crt.Build(ctx, req)
	return req.BuildOpts.Labels, err
}

func (st *Stack) evalCompEnv(comp *thrapb.Component, vm *pseudo.VM, scopeVars scope.Variables) error {
	for k, v := range comp.Env.Vars {
		result, err := vm.ParseEval(v, scopeVars)
		if err != nil {
			return err
		}
		if result.Type != hil.TypeString {
			return fmt.Errorf("env value must be string key=%s value=%s", k, v)
		}
		comp.Env.Vars[k] = result.Value.(string)
	}

	return nil
}

func (st *Stack) evalComponent(comp *thrapb.Component, scopeVars scope.Variables) error {

	var (
		vm  = pseudo.NewVM()
		err error
	)

	if comp.HasEnvVars() {
		err = st.evalCompEnv(comp, vm, scopeVars)
	}

	// TODO: In the future, eval other parts of the component

	return err
}
