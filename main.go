//author :Ally Dale(vipally@gmail.com)
//date: 2016-08-24

//tool installgithub is used to download the latest version of GitHub desktop offline install files
//refer: https://desktop.github.com
package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vipally/cmdline"
)

var (
	root_url    = "http://github-windows.s3.amazonaws.com" //下载根目录
	break_point = true
	local_root  = "."
	curl        = "./CURL.EXE"
	root_file   = "GitHub.application"
)

func main() {
	cmdline.Summary("Command installgithub is used to download the latest version of GitHub desktop offline install files")
	cmdline.Details("More information refer: https://desktop.github.com")
	cmdline.StringVar(&root_url, "r", "root", root_url, false, "root_url of GitHub desktop")
	cmdline.StringVar(&local_root, "d", "dir", ".", false, "local root dir for download")
	cmdline.BoolVar(&break_point, "b", "break_point", break_point, false, "if download from last break_point")
	file := cmdline.String("f", "file", "", false, "single file path to download")
	//*file = "Application Files\\GitHub_3_2_0_0\\GitHub.exe.manifest"
	cmdline.Parse()
	if *file != "" {
		dn_file(*file, false, -1)
		return
	} else {
		dn_from_root(break_point)
	}
}

type File struct {
	Type string
	Path string
	Size int
}

func get_dn_list(file string) (r []*File, err error) {
	full_path := local_dir(file)
	if f, err2 := os.Open(full_path); err == nil {
		d := xml.NewDecoder(f)
		for t, err3 := d.Token(); err3 == nil; t, err3 = d.Token() {
			switch token := t.(type) {
			case xml.StartElement:
				name := token.Name.Local
				if name == "dependentAssembly" {
					var nf File
					for _, attr := range token.Attr {
						switch attr.Name.Local {
						case "dependencyType":
							nf.Type = attr.Value
						case "codebase":
							nf.Path = attr.Value
						case "size":
							nf.Size, _ = strconv.Atoi(attr.Value)
						}
					}
					if nf.Type == "install" {
						r = append(r, &nf)
					}
				}
				if name == "file" {
					var nf File
					nf.Type = "install"
					for _, attr := range token.Attr {
						switch attr.Name.Local {
						case "name":
							nf.Path = attr.Value
						case "size":
							nf.Size, _ = strconv.Atoi(attr.Value)
						}
					}
					if nf.Type == "install" {
						r = append(r, &nf)
					}
				}
			case xml.EndElement:
			case xml.CharData:
			default:
			}
		}
	} else {
		err = err2
	}
	return
}

func dn_from_root(brk bool) error {
	dn_file(root_file, false, -1)
	if l, e := get_dn_list(root_file); e == nil {
		for _, v := range l {
			dn_file(v.Path, false, -1)
			dir := filepath.Dir(v.Path)
			if l2, e2 := get_dn_list(v.Path); e2 == nil {
				n := len(l2)
				size := 0.0
				for i, v2 := range l2 {
					v2.Path = dir + "\\" + v2.Path + ".deploy"
					s := float64(v2.Size) / 1024.0
					fmt.Printf("list%d/%d %.2fK: %s\n", i+1, n, s, v2.Path)
					size += s
				}
				size /= 1024.0
				fmt.Printf("total :%.2fM\n", size)
				dn := 0.0
				for i, v2 := range l2 {
					fmt.Printf("%d/%d %.2fM/%.2fM  %s\n", i+1, n, dn, size, v2.Path)
					dn_file(v2.Path, brk, v2.Size)
					dn += float64(v2.Size) / 1024.0 / 1024.0
				}
				fmt.Printf("\n\n\n!!!!!!!!!!!!!!Download finished, click [%s] to stat install!!!!!!!!!!!!!\n", root_file)
			} else {
				return e2
			}
		}
	} else {
		return e
	}
	return nil
}

func check_file(file string, size int) bool {
	if f, e := os.Open(file); e == nil {

		if s, e := f.Stat(); e == nil {
			if size <= 0 || s.Size() == int64(size) {
				f.Close()
				return true
			}
		}
		f.Close()
		err := os.Remove(file)
		fmt.Println("remove local file:", file, err)
	}
	return false
}

func dn_file(file string, brk bool, size int) error {
	url := full_url(file)
	local := local_dir(file)
	mk_dir(file)
	fmt.Println("downloding:", url)
	if brk && check_file(local, size) {
		fmt.Println("exist and skip", url)
	} else {
		cmd := exec.Command(curl, "-o", local, url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		cmd.Wait()
		fmt.Println("finish", url)
		if err != nil { //error then remove local file
			fmt.Println(err, url)
			fmt.Println(os.Remove(local))
		}
	}

	return nil
}

func full_url(file string) string {
	url := root_url + "/" + file
	url = strings.Replace(url, " ", "%20", -1)
	url = strings.Replace(url, "\\", "/", -1)
	return url
}
func local_dir(file string) string {
	return local_root + "/" + file
}

func mk_dir(file string) {
	d := filepath.Dir(file)
	os.MkdirAll(d, os.ModeDir)
}
