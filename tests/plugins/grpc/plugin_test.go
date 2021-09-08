package grpc

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	err := build()
	if err != nil {
		panic(err)
	}
}

func build() error {
	cmd := exec.Command("go", "build", "-o", "plugin", "../../../plugins/grpc/protoc_plugins/protoc-gen-php-grpc")
	return cmd.Run()
}

func protoc(t *testing.T, args []string) {
	cmd := exec.Command("protoc", "--plugin=protoc-gen-php-grpc=./plugin")
	cmd.Args = append(cmd.Args, args...)
	out, err := cmd.CombinedOutput()

	if len(out) > 0 || err != nil {
		t.Log("RUNNING: ", strings.Join(cmd.Args, " "))
	}

	if len(out) > 0 {
		t.Log(string(out))
	}

	if err != nil {
		t.Fatalf("protoc: %v", err)
	}
}

func Test_Simple(t *testing.T) {
	workdir, _ := os.Getwd()
	tmpdir, err := ioutil.TempDir("", "proto-test")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpdir))
	}()

	args := []string{
		"-Itestdata",
		"--php-grpc_out=" + tmpdir,
		"simple/simple.proto",
	}

	protoc(t, args)

	assertEqualFiles(
		t,
		workdir+"/testdata/simple/TestSimple/SimpleServiceInterface.php",
		tmpdir+"/TestSimple/SimpleServiceInterface.php",
	)
}

func Test_PhpNamespaceOption(t *testing.T) {
	workdir, _ := os.Getwd()
	tmpdir, err := ioutil.TempDir("", "proto-test")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpdir))
	}()

	args := []string{
		"-Itestdata",
		"--php-grpc_out=" + tmpdir,
		"php_namespace/service.proto",
	}
	protoc(t, args)

	assertEqualFiles(
		t,
		workdir+"/testdata/php_namespace/Test/CustomNamespace/ServiceInterface.php",
		tmpdir+"/Test/CustomNamespace/ServiceInterface.php",
	)
}

func Test_UseImportedMessage(t *testing.T) {
	workdir, _ := os.Getwd()
	tmpdir, err := ioutil.TempDir("", "proto-test")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpdir))
	}()

	args := []string{
		"-Itestdata",
		"--php-grpc_out=" + tmpdir,
		"import/service.proto",
	}
	protoc(t, args)

	assertEqualFiles(
		t,
		workdir+"/testdata/import/Import/ServiceInterface.php",
		tmpdir+"/Import/ServiceInterface.php",
	)
}

func Test_PhpNamespaceOptionInUse(t *testing.T) {
	workdir, _ := os.Getwd()
	tmpdir, err := ioutil.TempDir("", "proto-test")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpdir))
	}()

	args := []string{
		"-Itestdata",
		"--php-grpc_out=" + tmpdir,
		"import_custom/service.proto",
	}
	protoc(t, args)

	assertEqualFiles(
		t,
		workdir+"/testdata/import_custom/Test/CustomImport/ServiceInterface.php",
		tmpdir+"/Test/CustomImport/ServiceInterface.php",
	)
}

func Test_UseOfGoogleEmptyMessage(t *testing.T) {
	workdir, _ := os.Getwd()
	tmpdir, err := ioutil.TempDir("", "proto-test")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpdir))
	}()

	args := []string{
		"-Itestdata",
		"--php-grpc_out=" + tmpdir,
		"use_empty/service.proto",
	}
	protoc(t, args)

	assertEqualFiles(
		t,
		workdir+"/testdata/use_empty/Test/ServiceInterface.php",
		tmpdir+"/Test/ServiceInterface.php",
	)

	assert.NoError(t, os.RemoveAll("plugin"))
}

func assertEqualFiles(t *testing.T, original, generated string) {
	assert.FileExists(t, generated)

	originalData, err := ioutil.ReadFile(original)
	if err != nil {
		t.Fatal("Can't find original file for comparison")
	}

	generatedData, err := ioutil.ReadFile(generated)
	if err != nil {
		t.Fatal("Can't find generated file for comparison")
	}

	// every OS has a special boy
	r := strings.NewReplacer("\r\n", "", "\n", "")
	assert.Equal(t, r.Replace(string(originalData)), r.Replace(string(generatedData)))
}
