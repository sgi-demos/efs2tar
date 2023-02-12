// EFS .iso, .img, .efs => .tar
package main

import (
	"archive/tar"
	"errors"
	"flag"
	"log"
	"os"
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
	checkPath := flag.String("check", "", "the file to be checked for an efs filesystem")
	flag.Parse()

	checkEFSonly := len(*inputPath) == 0 && len(*checkPath) > 0
	if checkEFSonly {
		inputPath = checkPath
	}

	// allow input path as the only arg without specifying -in
	if len(*inputPath) == 0 {
		args := os.Args
		if len(*outputPath) == 0 && len(args) == 2 {
			*inputPath = args[1]
		} else {
			log.Fatal(errors.New("ERROR: need input EFS"))
		}
	}

	if !checkEFSonly {
		// generate tar file name if not provided
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

		if exists(*outputPath) {
			log.Fatal(errors.New("WARNING: tar exists: " + *outputPath))
		}
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
		log.Fatal(errors.New("WARNING: invalid EFS: " + *inputPath))
	} else {
		if !checkEFSonly {
			// write tar file
			outputFile, err := os.OpenFile(*outputPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
			if err != nil {
				log.Fatal(err)
			}
			tw := tar.NewWriter(outputFile)
			fs.WalkFilesystem(buildTarCallback(tw, fs))
			tw.Close()
			log.Fatal(errors.New("OK: valid EFS: " + *inputPath + "\nwrote tar: " + *outputPath))
		} else {
			// just check EFS
			log.Fatal(errors.New("OK: valid EFS: " + *inputPath))
		}
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
