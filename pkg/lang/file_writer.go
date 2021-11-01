package lang

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func WriteToFile(content string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer file.Close()

	file.WriteString(content)
}

func CompileWithKrakatau(filename string, outDir string) error {
	cmd := exec.Command("krakatau-assemble", filename)
	cmd.Dir = outDir

	fmt.Println(cmd.String())
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	fmt.Printf("%s\n", stdErr.String())
	fmt.Printf("%s\n", stdOut.String())
	return err
}
