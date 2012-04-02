package mogrify

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func init() {
	_, err := exec.LookPath("gm")
	if err != nil {
		log.Print("Could not locate gm (GraphicsMagic) tool in path")
		os.Exit(1)
	}
}

func Resize(out io.Writer, in io.Reader, size string) error {
	cmd := exec.Command("gm", "convert", "-resize", size, "-colorspace", "RGB", "-", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		log.Print(err)
		return err
	}

	// copy stream ot gm tool, close stream once done and 
	// read the output of the tool back to the out stream
	io.Copy(stdin, in)
	stdin.Close()
	io.Copy(out, stdout)

	return cmd.Wait()
}

func ResizeFile(filename string, size string) error {
	log.Printf("resize:%s", size)
	cmd := exec.Command("gm", "convert", "-resize", size, "-colorspace", "RGB", filename, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
  return cmd.Run()
}

