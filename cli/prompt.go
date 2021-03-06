/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/sniperkit/snk.fork.thrap/utils"
)

func promptForSupported(prompt string, supported []string, defaultVal string) string {
	prompt = prompt + " (" + strings.Join(supported, ", ") + ")"
	if defaultVal != "" {
		prompt += " [" + defaultVal + "]"
	}
	prompt += ": "

	val := defaultVal

	utils.PromptUntilNoError(prompt, os.Stdout, os.Stdin, func(db []byte) error {
		input := string(db)
		if input == "" && val != "" {
			return nil
		}

		for _, sd := range supported {
			if input == sd {
				val = input
				return nil
			}
		}

		return fmt.Errorf("not supported: '%s'", input)
	})

	return val
}

// func promptForSupported(userIn, prompt string, supported []string) string {
// 	prompt = prompt + " (" + strings.Join(supported, ", ") + "): "
//
// 	// Check if default input is supported
// 	for _, sd := range supported {
// 		if userIn == sd {
// 			fmt.Println(prompt + userIn)
// 			return userIn
// 		}
// 	}
//
// 	var selected string
// 	promptUntilNoError(prompt, os.Stdout, os.Stdin, func(db []byte) error {
// 		selected = string(db)
// 		for _, sd := range supported {
// 			if selected == sd {
// 				return nil
// 			}
// 		}
//
// 		return fmt.Errorf("not supported: %s", selected)
// 	})
//
// 	return selected
// }

// func promptUntilNoError(prompt string, out io.Writer, in io.Reader, f func([]byte) error) {
// 	var (
// 		lb []byte
// 	)
// 	err := io.ErrUnexpectedEOF
// 	for err != nil {
// 		out.Write([]byte(prompt))
//
// 		rd := bufio.NewReader(in)
// 		lb, _ = rd.ReadBytes('\n')
// 		lb = lb[:len(lb)-1]
// 		err = f(lb)
// 	}
// }
