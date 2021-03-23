package benchmarks

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/willabides/rjson"
)

func BenchmarkSkipValue(b *testing.B) {
	for _, sample := range benchSamples(b) {
		b.Run(sample.name, func(b *testing.B) {
			data := sample.data
			buffer := &rjson.Buffer{}
			var err error
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				benchInt, err = buffer.SkipValue(data)
			}
			require.NoError(b, err)
		})
	}
}

func BenchmarkGetValuesFromObject(b *testing.B) {
	type resType struct {
		PublicGists int64  `json:"public_gists"`
		PublicRepos int64  `json:"public_repos"`
		Login       string `json:"login"`
	}

	wantRes := resType{
		PublicGists: 8,
		PublicRepos: 8,
		Login:       "octocat",
	}

	data := []byte(exampleGithubUser)

	var res resType
	doneErr := fmt.Errorf("done")
	var err error
	buffer := &rjson.Buffer{}
	var stringBuf []byte
	var seenRepos, seenGists, seenLogin bool
	handler := rjson.ObjectValueHandlerFunc(func(fieldname, data []byte) (p int, err error) {
		switch string(fieldname) {
		case `public_gists`:
			res.PublicGists, p, err = rjson.ReadInt64(data)
			seenGists = true
		case `public_repos`:
			res.PublicRepos, p, err = rjson.ReadInt64(data)
			seenRepos = true
		case `login`:
			stringBuf, p, err = rjson.ReadStringBytes(data, stringBuf[:0])
			res.Login = string(stringBuf)
			seenLogin = true
		default:
			p, err = buffer.SkipValue(data)
		}
		if err == nil && seenGists && seenRepos && seenLogin {
			return p, doneErr
		}
		return p, err
	})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		seenGists, seenGists, seenLogin = false, false, false
		_, err = buffer.HandleObjectValues(data, handler)
	}
	require.Equal(b, wantRes, res)
	require.EqualError(b, err, "done")
}

var (
	benchInt  int
	benchBool bool
)

type benchSample struct {
	name string
	data []byte
}

func benchSamples(t testing.TB) []benchSample {
	return []benchSample{
		{
			name: "github user",
			data: []byte(exampleGithubUser),
		},
		{
			name: "large object",
			data: getTestdataJSONGz(t, "citm_catalog.json"),
		},
		{
			name: "unicode-heavy object",
			data: getTestdataJSONGz(t, "sample.json"),
		},
		{
			name: "string",
			data: []byte(`"this is a simple string"`),
		},
		{
			name: "integer value",
			data: []byte(`1234567`),
		},
		{
			name: "float value",
			data: []byte(`12.34567`),
		},
		{
			name: "null",
			data: []byte(`null`),
		},
	}
}

func gunzipTestJSON(t testing.TB, filename string) string {
	t.Helper()
	targetDir := filepath.Join("..", "testdata", "tmp")
	err := os.MkdirAll(targetDir, 0o700)
	require.NoError(t, err)
	target := filepath.Join(targetDir, filename)
	if fileExists(t, target) {
		return target
	}
	src := filepath.Join("..", "testdata", filename+".gz")
	f, err := os.Open(src)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()
	gz, err := gzip.NewReader(f)
	require.NoError(t, err)
	buf, err := ioutil.ReadAll(gz)
	require.NoError(t, err)
	err = ioutil.WriteFile(target, buf, 0o600)
	require.NoError(t, err)
	return target
}

func getTestdataJSONGz(t testing.TB, path string) []byte {
	t.Helper()
	filename := gunzipTestJSON(t, path)
	got, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	return got
}

func fileExists(t testing.TB, filename string) bool {
	t.Helper()
	_, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	require.NoError(t, err)
	return true
}

var exampleGithubUser = `{
 "avatar_url": "https://avatars.githubusercontent.com/u/583231?v=4",
 "bio": null,
 "blog": "https://github.blog",
 "company": "@github",
 "created_at": "2011-01-25T18:44:36Z",
 "email": "octocat@github.com",
 "events_url": "https://api.github.com/users/octocat/events{/privacy}",
 "followers": 3599,
 "followers_url": "https://api.github.com/users/octocat/followers",
 "following": 9,
 "following_url": "https://api.github.com/users/octocat/following{/other_user}",
 "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
 "gravatar_id": "",
 "hireable": null,
 "html_url": "https://github.com/octocat",
 "id": 583231,
 "location": "San Francisco",
 "login": "octocat",
 "name": "The Octocat",
 "node_id": "MDQ6VXNlcjU4MzIzMQ==",
 "organizations_url": "https://api.github.com/users/octocat/orgs",
 "public_gists": 8,
 "public_repos": 8,
 "received_events_url": "https://api.github.com/users/octocat/received_events",
 "repos_url": "https://api.github.com/users/octocat/repos",
 "site_admin": false,
 "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
 "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
 "twitter_username": null,
 "type": "User",
 "updated_at": "2021-03-22T14:27:47Z",
 "url": "https://api.github.com/users/octocat"
}`
