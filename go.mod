module github.com/sgi-demos/efs2tar

go 1.19

replace github.com/sgi-demos/efs2tar/efs => ./efs

replace github.com/sgi-demos/efs2tar/sgi => ./sgi

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/sgi-demos/efs2tar/efs v0.0.0-00010101000000-000000000000
	github.com/sgi-demos/efs2tar/sgi v0.0.0-00010101000000-000000000000
)
