package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/urfave/cli.v1"
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
	app.Version = "1.0.0"
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
			Usage: "Repository containing protocol buffer definitions of google apis, will be check out or updated under GOPATH",
			Value: "https://github.com/googleapis/googleapis",
		},
		cli.StringFlag{
			Name:  "proto-repo",
			Usage: "Repository containing core protocol buffer definitions from google, will be checked out or updated under GOPATH",
			Value: "https://github.com/google/protobuf",
		},
		cli.StringFlag{
			Name:  "dictybase-repo",
			Usage: "Repository containing protocol buffer definitions of dictybase api and services, will be checked out or updated under GOPATH",
			Value: "https://github.com/dictyBase/dictybaseapis",
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
	for _, cmd := range []string{"protoc", "protoc-gen-go", "protoc-gen-grpc-gateway"} {
		_, err := exec.LookPath(cmd)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("command %s not found %s", err),
				2,
			)
		}
	}
	return nil
}

func genProtoAction(c *cli.Context) error {
	log := getLogger(c)
	dictyDir, err := getFilePathFromRepo(c.String("dictybase-repo"))
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	if _, err := os.Stat(dictyDir); os.IsNotExist(err) {
		log.Debugf("repository %s does not exist at path %s, going to clone", c.String("dictybase-repo"), dictyDir)
		_, err = git.PlainClone(dictyDir, false, &git.CloneOptions{URL: c.String("dictybase-repo")})
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("unable to clone %s repo %s", dictyDir, err),
				2,
			)
		}
		log.Infof("cloned repository %s at path %s", c.String("dictybase-repo"), dictyDir)
	}
	apiDir, err := cloneOrUpdateGitRep(c.String("api-repo"), log)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	protoDir, err := cloneOrUpdateGitRep(c.String("proto-repo"), log)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
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
	//err = os.MkdirAll(output, 0775)
	//if err != nil {
	//return cli.NewExitError(
	//fmt.Sprintf("unable to create output folder %s %s", output, err),
	//2,
	//)
	//}
	for pkg, fnames := range pkgFiles {
		if !strings.HasPrefix(pkg, c.String("prefix")) {
			continue
		}
		includeDir := []string{apiDir, protoDir, dictyDir}
		includeDir = append(includeDir, filepath.Dir(fnames[0]))
		mapfn := func(path string) string {
			return filepath.Base(path)
		}
		names := Map(fnames, mapfn)
		if out, err := runProtoc(output, includeDir, names, log); err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in running protoc with output %s and error %s", string(out), err),
				2,
			)
			log.Infof(
				"ran protoc command on files %s with output %s",
				strings.Join(fnames, " "),
				string(out),
			)
		}
		// gateway plugin does not follow the package path, so
		// the exact path has to be given
		goutput := filepath.Join(os.Getenv("GOPATH"), "src", pkg)
		if out, err := runGrpcGateway(goutput, includeDir, names, log); err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in running protoc(grpc-gateway plugin) with output %s and error %s", string(out), err),
				2,
			)
			log.Infof(
				"ran protoc(grpc-gateway plugin) command on files %s with output %s",
				strings.Join(fnames, " "),
				string(out),
			)
		}
	}
	return nil
}

func getFilePathFromRepo(repo string) (string, error) {
	u, err := url.Parse(repo)
	if err != nil {
		return "", fmt.Errorf("unable to parse the given repository %s %s", repo, err)
	}
	// construct the full file path from repository
	path := filepath.Join(os.Getenv("GOPATH"), "src", u.Host, u.Path)
	return path, nil
}

func cloneOrUpdateGitRep(repo string, log *logrus.Logger) (string, error) {
	path, err := getFilePathFromRepo(repo)
	if err != nil {
		return "", err
	}
	// if the repository exists in the file system
	// update it
	if _, err := os.Stat(path); err == nil {
		log.Debugf("repository %s exist at path %s, going to update", repo, path)
		// get instance of a repository
		r, err := git.PlainOpen(path)
		if err != nil {
			return "", fmt.Errorf("unable to open repository %s", err)
		}
		// Get the working directory
		w, err := r.Worktree()
		if err != nil {
			return "", fmt.Errorf("unable to open work tree %s", err)
		}
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil {
			if err == git.NoErrAlreadyUpToDate {
				log.Infof("repository %s is uptodate", repo)
			} else {
				return "", fmt.Errorf("unable to pull %s", err)
			}
		} else {
			log.Infof("updated repository %s at path %s", repo, path)
		}
	} else { // clone the repository
		log.Debugf("repository %s does not exist at path %s, going to clone", repo, path)
		_, err = git.PlainClone(path, false, &git.CloneOptions{URL: repo})
		if err != nil {
			return "", fmt.Errorf("unable to clone %s repo %s", repo, err)
		}
		log.Infof("cloned repository %s at path %s", repo, path)
	}
	return path, nil
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
	args := []string{"--go_out=plugins=grpc:" + goOut}
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
