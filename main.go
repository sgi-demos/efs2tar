package main

import (
	"archive/tar"
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/sgi-demos/efs2tar/efs"
	"github.com/sgi-demos/efs2tar/sgi"
)

func exists(pathName string) bool {
	_, err := os.Stat(pathName)
	return !os.IsNotExist(err)
}

func main() {
	// get efs and tar file names
	inputPath := flag.String("in", "", "the file to be read as an efs filesystem")
	outputPath := flag.String("out", "", "the file to written to as a tar file")
	flag.Parse()
	if len(*inputPath) == 0 {
		log.Fatal(errors.New("ERROR: need at least an input filename"))
	}

	var outFile string
	var rootName string

	if len(*inputPath) > 0 && len(*outputPath) == 0 {
		inFile := *inputPath
		extPos := strings.LastIndex(inFile, ".")
		if extPos > 0 {
			rootName = inFile[:extPos]
		} else {
			rootName = inFile
		}
		outFile = rootName + ".tar"
		outputPath = &outFile
	}

	if _, err := os.Stat(*outputPath); !os.IsNotExist(err) {
		log.Fatal(errors.New("ERROR: output file already exists: " + *outputPath))
	}

	// read efs file
	file, err := os.Open(*inputPath)
	if err != nil {
		log.Fatal(err)
	}
	b := make([]byte, 51200)
	_, err = file.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	// ensure efs file is valid
	vh := sgi.NewVolumeHeader(b)
	p := vh.Partitions[7]
	fs := efs.NewFilesystem(file, p.Blocks, p.First)
	rootNode := fs.RootInode()
	if rootNode.Size == 0 {
		efsErrMsg := "invalid EFS file: " + *inputPath

		// on Mac, try attaching input file anyway, and if successful, tar it up
		if runtime.GOOS == "darwin" {
			log.Println("INFO:", errors.New(efsErrMsg))
			log.Println("INFO: trying: hdiutil attach / tar cvf / hdiutil detach")

			volName := "/Volumes/" + rootName
			existsBefore := exists(volName)
			if !existsBefore {
				cmd := exec.Command("hdiutil", "attach", *inputPath)
				cmd.Run()
				if exists(volName) {
					cmd = exec.Command("tar", "cvf", *outputPath, volName)
					cmd.Run()
					cmd = exec.Command("hdiutil", "detach", volName)
					cmd.Run()
					log.Fatal(errors.New("OK: valid non-EFS file: " + *inputPath))
				}
			}
			log.Fatal(errors.New("ERROR: invalid non-EFS file, or already mounted"))
		} else {
			log.Fatal(errors.New("ERROR: " + efsErrMsg))
		}
	} else {
		log.Fatal(errors.New("OK: valid EFS file: " + *inputPath))

		// write tar file
		outputFile, err := os.OpenFile(*outputPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal(err)
		}
		tw := tar.NewWriter(outputFile)
		fs.WalkFilesystem(buildTarCallback(tw, fs))
		tw.Close()
	}
}

func buildTarCallback(tw *tar.Writer, fs *efs.Filesystem) func(efs.Inode, string) {
	return func(in efs.Inode, path string) {
		if path == "" {
			return
		}

		if in.Type() == efs.FileTypeDirectory {
			hdr := &tar.Header{
				Name:     path,
				Mode:     0755,
				Typeflag: tar.TypeDir,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				log.Fatal(err)
			}
		} else if in.Type() == efs.FileTypeRegular {
			contents := fs.FileContents(in)
			hdr := &tar.Header{
				Name: path,
				Mode: 0755,
				Size: int64(len(contents)),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				log.Fatal(err)
			}
			if _, err := tw.Write([]byte(contents)); err != nil {
				log.Fatal(err)
			}
		} else if in.Type() == efs.FileTypeSymlink {
			contents := fs.FileContents(in)
			hdr := &tar.Header{
				Name:     path,
				Linkname: string(contents[:int64(len(contents))]),
				Typeflag: tar.TypeSymlink,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				log.Fatal(err)
			}
		}
	}
}
