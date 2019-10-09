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

	"github.com/sirupsen/logrus"
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
			Value: "https://github.com/google/protobuf",
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
			Value: "text",
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
	dictyDir := c.String("input-folder")
	apiDir, err := cloneGitRepo(c.String("api-repo"), "master")
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	log.Debugf("cloned repo %s at %s", c.String("api-repo"), apiDir)
	protoDir, err := cloneGitRepo(c.String("proto-repo"), "master")
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	log.Debugf("cloned repo %s at %s", c.String("proto-repo"), protoDir)
	protoDir = filepath.Join(protoDir, "src")
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
	if err := filepath.Walk(dictyDir, walkFn); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("%s directory walking errors %s", dictyDir, err),
			2,
		)
	}
	if len(output) == 0 {
		output = filepath.Join(os.Getenv("GOPATH"), "src")
	}
	for pkg, fnames := range pkgFiles {
		if !strings.HasPrefix(pkg, c.String("prefix")) {
			continue
		}
		// include cloned protocol buffer defintions
		includeDir := []string{apiDir, protoDir}
		// include the folder containing the protocol buffer files to compile
		includeDir = append(includeDir, dictyDir)
		// include the output folder
		includeDir = append(includeDir, output)
		// include the golang package folder dir as given in the proto defintion files
		includeDir = append(includeDir, filepath.Dir(fnames[0]))
		mapfn := func(path string) string {
			return filepath.Base(path)
		}
		// extract the protobuf file names from the full path
		names := Map(fnames, mapfn)
		out, err := runProtoc(output, includeDir, names, log)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in running protoc with output %s and error %s", string(out), err),
				2,
			)
		}
		log.Infof(
			"ran protoc command on files %s with output %s",
			strings.Join(fnames, " "),
			string(out),
		)
		// gateway plugin does not follow the package path, so
		// the exact path has to be given
		//goutput := filepath.Join(os.Getenv("GOPATH"), "src")
		//if out, err := runGrpcGateway(goutput, includeDir, names, log); err != nil {
		//return cli.NewExitError(
		//fmt.Sprintf("error in running protoc(grpc-gateway plugin) with output %s and error %s", string(out), err),
		//2,
		//)
		//log.Infof(
		//"ran protoc(grpc-gateway plugin) command on files %s with output %s",
		//strings.Join(fnames, " "),
		//string(out),
		//)
		//}
		//out, err := genProtoDocs(
		//goutput,
		//filepath.Join(goutput, c.String("prefix")),
		//includeDir,
		//names,
		//log,
		//)
		//if err != nil {
		//return cli.NewExitError(
		//fmt.Sprintf(
		//"error in running protoc(protoc-gen-doc plugin) with output %s and error %s",
		//string(out),
		//err,
		//),
		//2,
		//)
		//}
		//log.Infof(
		//"ran protoc(protoc-gen-doc plugin) command on files %s with output %s",
		//strings.Join(fnames, " "),
		//string(out),
		//)
		//if !c.Bool("swagger-gen") {
		//continue
		//}
		//if out, err := genSwaggerDefinition(c.String("swagger-output"), includeDir, names, log); err != nil {
		//return cli.NewExitError(
		//fmt.Sprintf("error in running protoc(swagger generator plugin) with output %s and error %s", string(out), err),
		//2,
		//)
		//log.Infof(
		//"ran protoc(swagger generator plugin) command on files %s with output %s",
		//strings.Join(fnames, " "),
		//string(out),
		//)
		//}
	}
	return nil
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
func runProtoc(goOut string, includes, fnames []string, log *logrus.Logger) ([]byte, error) {
	args := []string{
		"--go_out=plugins=grpc:" + goOut,
		"--govalidators_out=" + goOut,
	}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", filepath.Dir(fnames[0]))
	args = append(args, fnames...)
	log.Debugf("going to run protoc command %s", strings.Join(args, "\n"))
	return exec.Command("protoc", args...).CombinedOutput()
}

// runGrpcGateway executes "protoc" with grpc-gateway plugin on files named in fnames,
// passing go_out and include flags specified in goOut and includes respectively.
// It returns combined output from stdout and stderr.
func runGrpcGateway(goOut string, includes, fnames []string, log *logrus.Logger) ([]byte, error) {
	args := []string{"--grpc-gateway_out=allow_delete_body=true,logtostderr=true:" + goOut}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", filepath.Dir(fnames[0]))
	args = append(args, fnames...)
	log.Debugf("going to run protoc command %s", strings.Join(args, "\n"))
	return exec.Command("protoc", args...).CombinedOutput()
}

func genSwaggerDefinition(goOut string, includes, fnames []string, log *logrus.Logger) ([]byte, error) {
	args := []string{"--swagger_out=allow_delete_body=true,logtostderr=true:" + goOut}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", filepath.Dir(fnames[0]))
	args = append(args, fnames...)
	log.Debugf("going to run protoc command %s", strings.Join(args, "\n"))
	return exec.Command("protoc", args...).CombinedOutput()
}

func genProtoDocs(goOut, folder string, includes, fnames []string, log *logrus.Logger) ([]byte, error) {
	mdname := strings.Split(filepath.Base(fnames[0]), ".")[0]
	args := []string{
		fmt.Sprintf("--doc_out=%s", filepath.Join(folder, "docs")),
		fmt.Sprintf("--doc_opt=%s,%s.html", "html", mdname),
	}
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", filepath.Dir(fnames[0]))
	args = append(args, fnames...)
	log.Debugf("going to run protoc doc command %s", strings.Join(args, "\n"))
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
