/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package utils

import (
	"bufio"
	"io"
)

// PromptUntilNoError keeps prompting until a valid in put is supplied.  Input validity
// is confirmed by an input function
func PromptUntilNoError(prompt string, out io.Writer, in io.Reader, f func([]byte) error) {
	var (
		lb  []byte
		err = io.ErrUnexpectedEOF
	)

	for err != nil {
		out.Write([]byte(prompt))

		rd := bufio.NewReader(in)
		lb, _ = rd.ReadBytes('\n')
		lb = lb[:len(lb)-1]
		err = f(lb)
	}
}
