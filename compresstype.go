// compresstype is a package for detecting the compression format of a file. It also has functions to convert
//
// from one compression format to another.
package compresstype


import (
    "os/exec"
    "strings"
    "errors"
    "os"
    //"fmt"
    "path"
)

const (
    UNDEF int = iota
    PLAIN
    GZIP
    BZIP2
    XZ
    ZIP
)

func StringToType(t string) int {
    if t == "gzip" {
	return GZIP
    } else if t == "bzip2" {
	return BZIP2
    } else if t == "xz" {
	return XZ
    } else if t == "zip" {
	return ZIP
    }
    return PLAIN
}

func _extension(tp int) string {
    if tp == GZIP {
	return "gz"
    } else if tp == BZIP2 {
	return "bz2"
    } else if tp == XZ {
	return "xz"
    } else if tp == ZIP {
	return "zip"
    }
    return ""
}

func _compress_cmd(tp int) string {

    if tp == GZIP {
	return "gzip -f"
    } else if tp == BZIP2 {
	return "bzip2"
    } else if tp == XZ {
	return "xz"
    } else if tp == ZIP {
	return "zip -j"
    }
    return ""
}

func _uncompress_cmd(tp int) string {

    if tp == GZIP {
	return "gunzip"
    } else if tp == BZIP2 {
	return "bunzip2"
    } else if tp == XZ {
	return "unxz"
    } else if tp == ZIP {
	return "unzip -d"
    }
    return ""
}


func _parse_run_cmd(filepath, command string) (bool, string) {
    /*
     Parses command string and runs command on given filepath
     */
    parts := strings.Fields(command)
    env := []string{}
    cmd := []string{}
    for _, part := range parts {
	if strings.Index(part, "=") != -1 {
	    env = append(env, part)
	} else {
	    cmd = append(cmd, part)
	}
    }
    cmd = append(cmd, filepath)
    //fmt.Println("command : ", cmd)
    c := exec.Command(cmd[0], cmd[1:]...)
    c.Env = env
    out, _ := c.CombinedOutput()
    success := c.ProcessState.Success()
    return success, string(out)
}


func DetectFileType(name string) int {
    ok, out := _parse_run_cmd(name, "file")
    if ok == false {
	return UNDEF
    }
    l := len(name)
    //fmt.Println(out)
    if out[l+2:l+4] == "XZ" {
	return XZ
    } else if out[l+2:l+6] == "gzip" {
	return GZIP
    } else if out[l+2:l+7] == "bzip2" {
	return BZIP2
    } else if out[l+2:l+5] == "Zip" {
	return ZIP
    }
    return PLAIN
    
}

func Convert(srcfile, stringType string) (string, error) {
    destType := StringToType(stringType)
    srcType := DetectFileType(srcfile)
    if srcType == PLAIN {
	return _compress(srcfile, destType)
    }
    plain_file, err := _uncompress(srcfile, srcType)
    if err != nil {
	return plain_file, err
    }
    return _compress(plain_file, destType)
    
}

func _compress(srcfile string, destType int) (destfile string, err error) {
    destfile = srcfile + "." + _extension(destType)
    compress_cmd := _compress_cmd(destType)
    if destType == ZIP {
	compress_cmd = compress_cmd + " " + destfile
    }
    ok, out := _parse_run_cmd(srcfile, compress_cmd)
    if ok == false {
	return destfile, errors.New(out)
    }
    if destType == ZIP {
	os.Remove(srcfile)
    }
    return destfile, nil
}

func _uncompress(srcfile string, srcType int) (destfile string, err error) {
    parts := strings.Split(srcfile, ".")
    destfile = strings.Join(parts[0:len(parts)-1], ".")
    uncompress_command := _uncompress_cmd(srcType)
    if srcType == ZIP {
	// The file inside could have a different name
	ok, out := _parse_run_cmd(srcfile, "zipinfo -1")
	if ok == false {
	    return srcfile, errors.New(out)
	}
	out = strings.Trim(out, "\n\t\r ")
	if strings.Contains(out, "\n") == true {
	    return srcfile, errors.New("More than one file in the ZIP archive.This is not supported")
	}
	destfile = path.Join(path.Dir(srcfile), out)
	//fmt.Println("destfile = ", destfile)
	uncompress_command = uncompress_command + " " + path.Dir(srcfile)
    }
    ok, out := _parse_run_cmd(srcfile, uncompress_command)
    if ok == false {
	return destfile, errors.New(out)
    }
    if srcType == ZIP {
	os.Remove(srcfile)
    }
    return destfile, nil
}
