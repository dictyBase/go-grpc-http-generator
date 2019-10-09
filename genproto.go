package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var goPkgOptRe = regexp.MustCompile(`(?m)^option go_package = (.*);`)
var outputUsage = `
Base output path for generated source code.
		By default, it is $GOPATH/src along with the appended package path
`
var output string

func main() {
	app := cli.NewApp()
	app.Usage = "cli for generating go gRPC and gRPC-gateway source code for dictybase api and services"
	app.Name = "genproto"
	app.Version = "2.0.0"
	app.Author = "Siddhartha Basu"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "output,o",
			Usage:       outputUsage,
			Destination: &output,
		},
		cli.StringFlag{
			Name:  "prefix",
			Usage: "Go package prefix that will be matched to select definition files for code generation",
			Value: "github.com/dictyBase/go-genproto",
		},
		cli.StringFlag{
			Name:  "api-repo",
			Usage: "Repository containing protocol buffer definitions of google apis, will be check out",
			Value: "https://github.com/googleapis/googleapis",
		},
		cli.StringFlag{
			Name:  "proto-repo",
			Usage: "Repository containing core protocol buffer definitions from google, will be checked out",
			Value: "https://github.com/protocolbuffers/protobuf",
		},
		cli.StringFlag{
			Name:  "proto-repo-tag",
			Usage: "Repository tag for protocol buffer repo",
			Value: "v3.9.2",
		},
		cli.StringFlag{
			Name:  "validator-repo",
			Usage: "Repository containing protocol buffer definitions for validation, will be checked out",
			Value: "https://github.com/mwitkow/go-proto-validators.git",
		},
		cli.StringFlag{
			Name:  "validator-repo-tag",
			Usage: "Repository tag for validation protocol buffer",
			Value: "v0.2.0",
		},
		cli.StringFlag{
			Name:  "input-folder,i",
			Usage: "Folder containing protocol buffer definitions, will be looked up recursively",
		},
		cli.StringFlag{
			Name:  "log-level",
			Usage: "log level for the application",
			Value: "error",
		},
		cli.StringFlag{
			Name:  "log-format",
			Usage: "format of the logging out, either of json or text",
			Value: "json",
		},
		cli.BoolFlag{
			Name:  "swagger-gen",
			Usage: "generate swagger definition files from grpc-gateway definition",
		},
		cli.StringFlag{
			Name:  "swagger-output",
			Usage: "Output folder for swagger definition files, should be set with swagger-gen option",
		},
	}
	app.Before = validateGenProto
	app.Action = genProtoAction
	app.Run(os.Args)
}

func validateGenProto(c *cli.Context) error {
	// check if GOPATH is defined
	_, ok := os.LookupEnv("GOPATH")
	if !ok {
		return cli.NewExitError("GOPATH env is not defined", 2)
	}
	// check if all the required binaries are in path
	for _, cmd := range []string{"protoc", "protoc-gen-go", "protoc-gen-grpc-gateway", "protoc-gen-doc"} {
		_, err := exec.LookPath(cmd)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("command %s not found %s", cmd, err),
				2,
			)
		}
	}
	if c.Bool("swagger-gen") {
		if len(c.String("swagger-output")) == 0 {
			return cli.NewExitError(
				"*swagger-ouput* have to be set for *swagger-gen* to work",
				2,
			)
		}
	}
	return nil
}

func genProtoAction(c *cli.Context) error {
	log := getLogger(c)
	if len(output) == 0 {
		output = filepath.Join(os.Getenv("GOPATH"), "src")
	}
	dictyDir := c.String("input-folder")

	pkgFiles, err := mapPath2Proto(dictyDir)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("%s directory walking errors %s", dictyDir, err),
			2,
		)
	}

	valProtoDir, err := cloneGitTag(c.String("validator-repo"), c.String("validator-repo-tag"))
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("error in cloning repo %s %s", c.String("validator-repo"), err),
			2,
		)
	}
	log.Debugf("cloned repo %s at %s", c.String("validator-repo"), valProtoDir)
	apiDir, err := cloneGitRepo(c.String("api-repo"), "master")
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	log.Debugf("cloned repo %s at %s", c.String("api-repo"), apiDir)
	protoDir, err := cloneGitTag(c.String("proto-repo"), c.String("proto-repo-tag"))
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	log.Debugf("cloned repo %s at %s", c.String("proto-repo"), protoDir)
	protoDir = filepath.Join(protoDir, "src")

	// include
	//	i)   cloned protocol buffer defintions
	//	ii)  folder containing the protocol buffer files to compile
	//	iii) validate proto files
	//	iv)  output folder
	baseIncludeDir := []string{
		apiDir,
		protoDir,
		dictyDir,
		valProtoDir,
		output,
	}
	for pkg, fnames := range pkgFiles {
		if !strings.HasPrefix(pkg, c.String("prefix")) {
			continue
		}
		var includeDir []string
		// include the golang package folder dir as given in the proto defintion files
		includeDir = append(baseIncludeDir, filepath.Dir(fnames[0]))
		// extract the protobuf file names from the full path
		names := Map(fnames, func(path string) string {
			return filepath.Base(path)
		})
		out, err := runProtoc(output, includeDir, names)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in running protoc with output %s and error %s", string(out), err),
				2,
			)
		}
		log.Debugf(
			"ran protoc command on files %s with output %s",
			strings.Join(fnames, " "),
			string(out),
		)
		// gateway plugin does not follow the package path, so
		// the exact path has to be given
		//goutput := filepath.Join(os.Getenv("GOPATH"), "src")
		out, err = runGrpcGateway(output, includeDir, names)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in running protoc(grpc-gateway plugin) with output %s and error %s", string(out), err),
				2,
			)
		}
		log.Infof(
			"ran protoc(grpc-gateway plugin) command on files %s with output %s",
			strings.Join(fnames, " "),
			string(out),
		)

		out, err = genProtoDocs(
			output,
			filepath.Join(output, c.String("prefix")),
			includeDir,
			names,
		)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf(
					"error in running protoc(protoc-gen-doc plugin) with output %s and error %s",
					string(out),
					err,
				),
				2,
			)
		}
		log.Debugf(
			"ran protoc(protoc-gen-doc plugin) command on files %s with output %s",
			strings.Join(fnames, " "),
			string(out),
		)
		if !c.Bool("swagger-gen") {
			continue
		}
		out, err = genSwaggerDefinition(c.String("swagger-output"), includeDir, names)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf(
					"error in running protoc(swagger generator plugin) with output %s and error %s",
					string(out), err,
				),
				2,
			)
		}
		log.Debugf(
			"ran protoc(swagger generator plugin) command on files %s with output %s",
			strings.Join(fnames, " "),
			string(out),
		)
	}
	return nil
}

func cloneGitTag(repo, tag string) (string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "gclone")
	if err != nil {
		return "", fmt.Errorf("error in creating temp dir %s", err)
	}
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:           repo,
		SingleBranch:  true,
		ReferenceName: plumbing.NewTagReferenceName(tag),
	})
	return dir, err
}

func cloneGitRepo(repo, branch string) (string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "gclone")
	if err != nil {
		return "", fmt.Errorf("error in creating temp dir %s", err)
	}
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:           repo,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	return dir, err
}

// goPkg reports the import path declared in the given file's
// `go_package` option. If the option is missing, goPkg returns empty string.
func goPkg(fname string) (string, error) {
	content, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}

	var pkgName string
	if match := goPkgOptRe.FindSubmatch(content); len(match) > 0 {
		pn, err := strconv.Unquote(string(match[1]))
		if err != nil {
			return "", err
		}
		pkgName = pn
	}
	if p := strings.IndexRune(pkgName, ';'); p > 0 {
		pkgName = pkgName[:p]
	}
	return pkgName, nil
}

// runProtoc executes the "protoc" command on files named in fnames,
// passing go_out and include flags specified in goOut and includes respectively.
// protoc returns combined output from stdout and stderr.
func runProtoc(goOut string, includes, fnames []string) ([]byte, error) {
	args := []string{
		"--go_out=plugins=grpc:" + goOut,
		"--govalidators_out=" + goOut,
	}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", filepath.Dir(fnames[0]))
	args = append(args, fnames...)
	return exec.Command("protoc", args...).CombinedOutput()
}

// runGrpcGateway executes "protoc" with grpc-gateway plugin on files named in fnames,
// passing go_out and include flags specified in goOut and includes respectively.
// It returns combined output from stdout and stderr.
func runGrpcGateway(goOut string, includes, fnames []string) ([]byte, error) {
	args := []string{"--grpc-gateway_out=allow_delete_body=true,logtostderr=true:" + goOut}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", filepath.Dir(fnames[0]))
	args = append(args, fnames...)
	return exec.Command("protoc", args...).CombinedOutput()
}

func genSwaggerDefinition(goOut string, includes, fnames []string) ([]byte, error) {
	args := []string{"--swagger_out=allow_delete_body=true,logtostderr=true:" + goOut}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", filepath.Dir(fnames[0]))
	args = append(args, fnames...)
	return exec.Command("protoc", args...).CombinedOutput()
}

func genProtoDocs(goOut, folder string, includes, fnames []string) ([]byte, error) {
	doc := filepath.Join(folder, "docs")
	if err := os.MkdirAll(doc, 0775); err != nil {
		return []byte{}, err
	}
	mdname := strings.Split(filepath.Base(fnames[0]), ".")[0]
	args := []string{
		fmt.Sprintf("--doc_out=%s", doc),
		fmt.Sprintf("--doc_opt=%s,%s.html", "html", mdname),
	}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", filepath.Dir(fnames[0]))
	args = append(args, fnames...)
	return exec.Command("protoc", args...).CombinedOutput()
}

// Map applies the given function to each element of a, returning slice of
// results
func Map(a []string, fn func(string) string) []string {
	if len(a) == 0 {
		return a
	}
	sl := make([]string, len(a))
	for i, v := range a {
		sl[i] = fn(v)
	}
	return sl
}

// mapPath2Proto maps go package import path to their corresponding proto files
func mapPath2Proto(path string) (map[string][]string, error) {
	pkgFiles := make(map[string][]string)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || !strings.HasSuffix(path, ".proto") {
			return nil
		}
		pkg, err := goPkg(path)
		if err != nil {
			return err
		}
		pkgFiles[pkg] = append(pkgFiles[pkg], path)
		return nil
	}
	if err := filepath.Walk(path, walkFn); err != nil {
		return pkgFiles, err
	}
	return pkgFiles, nil
}
