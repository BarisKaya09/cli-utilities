package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Cli struct {
	*CommandParser
	name   Commands
	params any
}

func (c *Cli) Echo() error {
	// 	echo: repeats input (tekrar eder)
	fmt.Println(c.params)
	return nil
}

func (c *Cli) Cat() error {
	// cat: concatenates files (parametre olarak verilen dosya ve ya dosyaların içeriklerini yazdırır)
	contentch := make(chan string)
	quitch := make(chan error)
	wg := sync.WaitGroup{}
	for _, file := range c.params.([]string) {
		wg.Add(1)
		go func(f *string) {
			defer wg.Done()
			content, err := os.ReadFile(*f)
			if err != nil {
				select {
				case quitch <- err:
				default:
				}
				return
			}
			contentch <- string(content)
		}(&file)
	}

	go func() {
		wg.Wait()
		close(contentch)
	}()

	select {
	case err := <-quitch:
		return err
	case <-time.After(time.Second):
		fmt.Println("Timeout occurred")
	}

	for content := range contentch {
		fmt.Println(content)
	}

	return nil
}

func (c *Cli) Ls() error {
	// ls: lists directories (dizini listeler)
	if err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var dtype string
		if info.IsDir() {
			dtype = "directory"
		} else {
			dtype = "file"
		}
		fmt.Printf("name: %v type: %v\n", path, dtype)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (c *Cli) Find() error {
	// find: locates files or directories (dizinde dosya veya dizin arar)
	p := c.params
	var founded string
	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == p {
			if info.IsDir() {
				dir, err := os.ReadDir(path)
				if err != nil {
					return err
				}
				founded = fmt.Sprintf("%v\n", path)
				for _, d := range dir {
					founded += fmt.Sprintf("  -> %v\n", d)
				}
			} else {
				founded = path
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	if len(founded) == 0 {
		return errors.New("not found file or directory")
	}

	fmt.Println(founded)

	return nil
}

func (c *Cli) Grep() error {
	// grep: matches text in files (dosyalarda metin arar)
	fileName := c.params.([]string)[1]
	substr := c.params.([]string)[0]
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	contentScanner := bufio.NewScanner(file)
	foundedCount := 0
	for contentScanner.Scan() {
		if strings.Contains(contentScanner.Text(), substr) {
			foundedCount++
			fmt.Printf("%v\n", contentScanner.Text())
		}
	}
	fmt.Printf("founded count %v", foundedCount)
	return nil
}

func (c *Cli) Run() {
	if err := c.Parse(c); err != nil {
		log.Fatalf("parse error: %v", err)
	}

	switch c.name {
	case Echo:
		// 	echo: repeats input (tekrar eder)
		if err := c.Echo(); err != nil {
			log.Fatalf("echo error: %v", err)
		}
	case Cat:
		// cat: concatenates files (parametre olarak verilen dosya ve ya dosyaların içeriklerini yazdırır)
		if err := c.Cat(); err != nil {
			log.Fatalf("cat error: %v", err)
		}
	case Ls:
		// ls: lists directories (dizini listeler)
		if err := c.Ls(); err != nil {
			log.Fatalf("ls error: %v", err)
		}
	case Find:
		// find: locates files or directories (dizinde dosya veya dizin arar)
		if err := c.Find(); err != nil {
			log.Fatalf("find error: %v", err)
		}
	case Grep:
		// grep: matches text in files (dosyalarda metin arar)
		if err := c.Grep(); err != nil {
			log.Fatalf("grep error: %v", err)
		}
	default:
	}
}

// burada işler farklı

type Commands string

const (
	Echo Commands = "echo"
	Cat  Commands = "cat"
	Ls   Commands = "ls"
	Find Commands = "find"
	Grep Commands = "grep"
)

type CommandParser struct{}

func (cp *CommandParser) Parse(cli *Cli) error {
	args := os.Args[1:]
	if len(args) == 0 {
		return errors.New("not found args")
	}

	switch Commands(args[0]) {
	case Echo:
		// 	echo: repeats input (tekrar eder)
		if len(args[1:]) == 0 {
			return errors.New("invalid params length for echo")
		}
		cli.name = Echo
		cli.params = strings.Join(args[1:], " ")
	case Cat:
		// cat: concatenates files (parametre olarak verilen dosya ve ya dosyaların içeriklerini yazdırır)
		if len(args[1:]) == 0 {
			return errors.New("invalid params length for cat")
		}
		cli.name = Cat
		cli.params = args[1:]
	case Ls:
		// ls: lists directories (dizini listeler)
		if len(args) > 1 {
			return errors.New("invalid args length for ls")
		}
		cli.name = Ls
		cli.params = nil
	case Find:
		// find: locates files or directories (dizinde dosya veya dizin arar)
		if len(args[1:]) != 1 {
			return errors.New("invalid params length for find")
		}
		cli.name = Find
		cli.params = args[1]
	case Grep:
		// grep: matches text in files (dosyalarda metin arar)
		if len(args[1:]) != 2 {
			return errors.New("invalid params length for grep")
		}
		cli.name = Grep
		cli.params = args[1:]
	default:
		return errors.New("invalid command")
	}

	return nil
}

func main() {
	cli := Cli{}
	cli.Run()
}
