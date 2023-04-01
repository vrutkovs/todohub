package settings

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/vrutkovs/todohub/pkg/source/github"
	"github.com/vrutkovs/todohub/pkg/storage/todoist"
	"github.com/vrutkovs/todohub/pkg/storage/trello"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSettingsInterface(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Settings")
}

type FakeReadFiler struct {
	Str string
}

func (f FakeReadFiler) fakeReadFile(filename string) ([]byte, error) {
	buf := bytes.NewBufferString(f.Str)
	return ioutil.ReadAll(buf)
}

var errFileRead = errors.New("no such file")

func (f FakeReadFiler) fakeReadFileThrowsError(filename string) ([]byte, error) {
	return []byte{}, errFileRead
}

func mockSettings(s Settings) (string, error) {
	data, err := yaml.Marshal(s)
	if err != nil {
		return bytes.NewBuffer([]byte{}).String(), err
	}
	return bytes.NewBuffer(data).String(), nil
}

var _ = DescribeTable("LoadSettings",
	func(settings Settings) {
		data, err := mockSettings(settings)
		Expect(err).NotTo(HaveOccurred())
		f := FakeReadFiler{
			Str: data,
		}
		s, err := LoadSettings("/dev/null", f.fakeReadFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).NotTo(Equal(settings))
	},
	Entry("Empty", Settings{}),
	Entry("Trello", Settings{
		Storage: StorageSettings{
			Trello: &trello.Settings{
				Token: "foobar",
			},
		},
	},
	),
	Entry("Todoist", Settings{
		Storage: StorageSettings{
			Todoist: &todoist.Settings{
				Token: "foobar",
			},
		},
	},
	),
	Entry("Github", Settings{
		Source: SourceSettings{
			Github: &github.Settings{
				Token:      "foobar",
				SearchList: map[string]string{},
			},
		},
	},
	),
)

var _ = Describe("LoadSettings: errors", func() {
	It("wraps read error", func() {
		data, err := mockSettings(Settings{})
		Expect(err).NotTo(HaveOccurred())
		f := FakeReadFiler{
			Str: data,
		}
		s, err := LoadSettings("/no/such/file", f.fakeReadFileThrowsError)
		Expect(err).To(MatchError(errFileRead))
		Expect(s).To(BeNil())
	})

	It("wraps unmarshal error", func() {
		data := `{"name":what?}`
		f := FakeReadFiler{
			Str: data,
		}
		s, err := LoadSettings("/dev/null", f.fakeReadFile)
		Expect(err).To(MatchError(errors.New("yaml: did not find expected ',' or '}'")))
		Expect(s).To(BeNil())
	})
})

type trelloError struct {
	msg  string
	code int
}

func (e trelloError) Error() string {
	return e.msg
}

var (
	trelloSettings = trello.Settings{
		AppKey:  "token",
		Token:   "trello",
		BoardID: "trello",
	}
	todoistSettings = todoist.Settings{
		Token: "todoist",
	}
	trello401Error = trelloError{msg: "HTTP request failure on https://api.trello.com/1/boards/trello:\n401: invalid key", code: 401}
)

var _ = DescribeTable("GetActiveStorageClient",
	func(storage StorageSettings, targetErr error) {
		_, err := storage.GetActiveStorageClient()
		Expect(err.Error()).To(Equal(targetErr.Error()))
	},
	Entry("Empty", StorageSettings{}, fmt.Errorf("no valid storage settings found")),
	Entry("Trello", StorageSettings{
		Trello: &trelloSettings,
	}, trello401Error),
	Entry("Todoist", StorageSettings{
		Todoist: &todoistSettings,
	}, errors.New("failed to sync, status code: 403, command: []")),
	Entry("Both", StorageSettings{
		Trello:  &trelloSettings,
		Todoist: &todoistSettings,
	}, trello401Error),
)
