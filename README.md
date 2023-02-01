# efs2tar

This is a fork of https://github.com/sophaskins/efs2tar.  `efs2tar` is a tool that converts SGI EFS-formatted filesystem images into tarballs. It was based entirely on NetBSD's `sys/fs/efs` ([source](http://cvsweb.netbsd.org/bsdweb.cgi/src/sys/fs/efs/?only_with_tag=MAIN)).

The goal of this fork is to streamline the command line, in particular to auto generate output file names so that it can be run batched.  For example, `find . -name "*.iso" -print -exec efs2tar {} ;\`.  The auto-generated tarball name is generated simply by replacing the source file's extension with `.tar`. 

## Example usage

```
$ go install github.com/sgi-demos/efs2tar
$ efs2tar -in ~/my-sgi-disc.iso -out ~/my-sgi-disc.tar
```

The Internet Archive has [several discs](https://archive.org/search.php?query=sgi&and%5B%5D=mediatype%3A%22software%22&page=2) in its collections that are formatted with EFS.


## "Edge cases" not covered
* any type of file other than directories and normal files (which is to say, links in particular do not work)
* partition layouts other than what you'd expect to see on an SGI-produced CDROM
* any sort of error handling...at all
* that includes verifying magic numbers
* preserving the original file permissions
* I've only tested this on like, one CD