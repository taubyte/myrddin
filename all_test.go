package myrddin

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"archive/tar"
	"archive/zip"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func init() {
	ProcessingFileName = func() string {
		return fmt.Sprintf("config-test-%d.debug", time.Now().UnixNano())
	}
}

var (
	envYamlFixture = map[string]interface{}{
		"var1": "{{ hostname }}",
		"var2": 42,
		"var3": 3.14,
		"var4": struct {
			name string
			desc string
		}{
			name: "a name",
			desc: "some desc",
		},
		"var5": "{{ .version }}",
		"var6": "some value",
	}
	yamlFixtures = map[string]string{
		"part1.yaml": `
Section1:
  val1: 1
  val2: &val2 {{env "var6"}}`,
		"part2.yaml": `
Section2:
  val1: "{{ hostname }}"
  val2: 3.14
  val3: *val2
    `}
	yamlFixtures2 = map[string]string{
		"part3.yaml": `
Section3:
  val1: {{ me "Test" 0}}
  val2: {{ .who }}
    `}

	base64yamlFixture = map[string]string{
		"part4.yaml": `
Section4:
    val1: {{ based "./all_test.go" }}
    `}
)

type section1Struct struct {
	Val1 int    `yaml:"val1"`
	Val2 string `yaml:"val2"`
}

type configStruct struct {
	Section1 section1Struct         `yaml:"Section1"`
	Section2 map[string]interface{} `yaml:"Section2"`
	Section3 map[string]string      `yaml:"Section3"`
	Section4 map[string]string      `yaml:"Section4"`
}

func fixture_env_yaml(main_fs afero.Fs) error {
	// create env yaml file
	path := EnvironementFileName
	env_yaml, err := main_fs.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(0640))
	if err != nil {
		return err
	}
	defer env_yaml.Close()

	env_data, err := yaml.Marshal(envYamlFixture)
	if err != nil {
		return err
	}

	_, err = env_yaml.Write(env_data)
	if err != nil {
		return err
	}

	// print content
	f, _ := main_fs.Open(path)
	ioutil.ReadAll(f)

	return nil
}

func fixture_yaml(main_fs afero.Fs, fixture map[string]string) error {
	for fname, fdata := range fixture {
		// create env yaml file
		f_yaml, err := main_fs.OpenFile(fname, os.O_CREATE|os.O_WRONLY, os.FileMode(0640))
		if err != nil {
			return err
		}
		defer f_yaml.Close()

		_, err = f_yaml.WriteString(fdata)
		if err != nil {
			return err
		}
	}
	return nil
}

func fixtures(stage int) (afero.Fs, error) {
	main_fs := afero.NewMemMapFs()
	if stage == 1 {
		return main_fs, nil
	}

	err := fixture_env_yaml(main_fs)
	if err != nil {
		return nil, err
	}

	if stage == 2 {
		return main_fs, nil
	}

	err = fixture_yaml(main_fs, yamlFixtures)
	if err != nil {
		return nil, err
	}

	if stage == 3 {
		return main_fs, nil
	}

	err = fixture_yaml(main_fs, yamlFixtures2)
	if err != nil {
		return nil, err
	}

	return main_fs, nil
}

func TestLoadingEnvYaml(t *testing.T) {
	fs, err := fixtures(2)
	if err != nil {
		t.Error(err)
		return
	}

	m, _ := New(nil)

	m.store = fs

	err = m.Environement().parseEnvironement()
	if err != nil {
		t.Error(err)
		return
	}

	{ // check template functions are working
		v, e1 := m.env.Get("var1")
		h, e2 := os.Hostname()
		if e1 != nil || e2 != nil || v.(string) != h {
			t.Error("Fail to process template: Did not subsitute hostname!")
		}
	}

	{ // check template data is working
		v, e1 := m.env.Get("var5")
		if e1 != nil || v.(string) != Version {
			t.Error("Fail to process template: Did not subsitute data!")
		}
	}
}

func TestParseAllYamlWoOptions(t *testing.T) {
	fs, err := fixtures(3)
	if err != nil {
		t.Error(err)
		return
	}

	config := configStruct{}

	m, err := New(&config)
	if err != nil {
		t.Error(err)
		return
	}

	m.store = fs

	err = m.Parse()
	if err != nil {
		t.Error(err)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	if _, ok := m.config.(*configStruct); fmt.Sprint(m.config) != fmt.Sprintf("&{{1 some value} map[val1:%s val2:3.14 val3:some value] map[] map[]}", hostname) || ok == false {
		t.Error("Failed to parse all YAML files")
	}

}

func TestParseAllYamlWOptions(t *testing.T) {
	fs, err := fixtures(4)
	if err != nil {
		t.Error(err)
		return
	}

	config := configStruct{}

	m, _ := New(
		&config,
		Function("me", func(name string, num int) string { return fmt.Sprintf("%s=%d?", name, num) }),
		Data("who", "the Mage!"),
	)

	m.store = fs

	err = m.Parse()
	if err != nil {
		t.Error(err)
		return
	}

	if cnf := m.config.(*configStruct); fmt.Sprint(cnf.Section3) != "map[val1:Test=0? val2:the Mage!]" {
		t.Error("Failed to parse all YAML files")
	}
}

/*
image.png.base64
opensssl base64 -d < image.png.base64 > image.png
*/
func TestParseBase64(t *testing.T) {
	main_fs := afero.NewMemMapFs()
	fixture_env_yaml(main_fs)
	err := fixture_yaml(main_fs, base64yamlFixture)
	if err != nil {
		t.Error(err)
		return
	}

	config := configStruct{}

	m, _ := New(
		&config,
		Function("based", func(filename string) string {
			byt, err := os.ReadFile(filename)
			if err != nil {
				return "ERROR"
			}
			return base64.StdEncoding.EncodeToString(byt)
		}),
	)

	m.store = main_fs

	err = m.Parse()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestExampleDirectory(t *testing.T) {
	main_fs := afero.NewBasePathFs(afero.NewOsFs(), "./examples/improved")

	fixture_env_yaml(main_fs)
	f_yaml, err := main_fs.OpenFile("/index.yaml", os.O_CREATE|os.O_WRONLY, os.FileMode(0640))
	if err != nil {
		t.Error(err)
		return
	}
	defer f_yaml.Close()

	var config interface{}

	m, err := New(
		&config,
		DefaultLoaders(),
	)
	if err != nil {
		t.Error(err)
		return
	}

	m.store = main_fs

	err = m.Parse()
	if err != nil {
		t.Error(err)
		return
	}
}

/// ZIP

func createZip(main_fs afero.Fs) (string, error) {
	tf, err := ioutil.TempFile("", "tb-myrddin-*.zip")
	if err != nil {
		return "", err
	}

	zw := zip.NewWriter(tf)
	defer zw.Close()

	err = afero.Walk(main_fs, "/", func(path string, info fs.FileInfo, err error) error {
		if info == nil || info.IsDir() == true {
			return nil
		}

		w, err := zw.Create(path[1:])
		if err != nil {
			return err
		}

		r, _ := main_fs.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()

		data, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		_, err = w.Write(data)
		if err != nil {
			return err
		}

		return nil
	})

	return tf.Name(), err

}

func TestZipLoad(t *testing.T) {
	fs, err := fixtures(3)
	if err != nil {
		t.Error(err)
		return
	}

	path, err := createZip(fs)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(path)

	config := configStruct{}

	m, _ := New(&config)

	err = m.Load("file://" + path)
	if err != nil {
		t.Error(err)
		return
	}

	err = m.Parse()
	if err != nil {
		t.Error(err)
		return
	}

}

/// TAR

func createTar(main_fs afero.Fs) (string, error) {
	tf, err := ioutil.TempFile("", "tb-myrddin-*.tar")
	if err != nil {
		return "", err
	}

	tw := tar.NewWriter(tf)
	defer tw.Close()

	err = afero.Walk(main_fs, "/", func(path string, info fs.FileInfo, err error) error {
		if info == nil || info.IsDir() == true {
			return nil
		}

		r, _ := main_fs.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()

		header := &tar.Header{
			Name:    path[1:],
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("Could not write header for file '%s', got error '%w'", path, err)
		}

		_, err = io.Copy(tw, r)
		if err != nil {
			return fmt.Errorf("Could not copy the file '%s' data to the tarball, got error '%w'", path, err)
		}

		return nil
	})

	// check tar
	tf.Seek(0, 0)
	t := tar.NewReader(tf)
	for {
		_, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		_, err = buf.ReadFrom(t)
		if err != nil {
			return "", err
		}

	}

	return tf.Name(), err

}

func TestTarLoad(t *testing.T) {
	fs, err := fixtures(3)
	if err != nil {
		t.Error(err)
		return
	}

	path, err := createTar(fs)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(path)

	config := configStruct{}

	m, _ := New(&config)

	err = m.Load("file://" + path)
	if err != nil {
		t.Error(err)
		return
	}

	err = m.Parse()
	if err != nil {
		t.Error(err)
		return
	}

}
