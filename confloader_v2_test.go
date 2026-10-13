package confloader_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	cnf "github.com/sizzlei/confloader"
)

const testRegion = "ap-northeast-2"

const testParamYAML = `Param:
  - ConfigId: repository
    Conf:
      Host: db.local
      Port: 3306
      DB:
        - argus
        - lynx
  - ConfigId: server
    Conf:
      Listen: ":8080"
`

// capturedRequest 는 가짜 서버가 받은 요청 중 검증에 필요한 값입니다.
type capturedRequest struct {
	Target        string
	Authorization string
	Body          map[string]interface{}
}

// fakeAWS 는 SSM·Secrets Manager 가 쓰는 JSON 1.1 프로토콜을 흉내 내는 테스트 서버입니다.
// 요청 하나를 받아 기록하고 status·resp 로 응답합니다.
func fakeAWS(t *testing.T, status int, resp string) (*httptest.Server, *capturedRequest) {
	t.Helper()

	got := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Target = r.Header.Get("X-Amz-Target")
		got.Authorization = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&got.Body); err != nil {
			t.Errorf("요청 본문 해석 실패: %v", err)
		}

		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(resp))
	}))
	t.Cleanup(srv.Close)

	return srv, got
}

// testOpts 는 실제 AWS 대신 가짜 서버로 요청하도록 하는 로드 옵션입니다.
// 로컬 ~/.aws 설정이나 AWS_PROFILE 이 결과에 끼어들지 않도록 환경도 비웁니다.
func testOpts(t *testing.T, endpoint string) []func(*config.LoadOptions) error {
	t.Helper()

	empty := filepath.Join(t.TempDir(), "none")
	t.Setenv("AWS_CONFIG_FILE", empty)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", empty)
	t.Setenv("AWS_PROFILE", "")

	return []func(*config.LoadOptions) error{
		config.WithBaseEndpoint(endpoint),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("AKIDTEST", "SECRETTEST", "")),
		config.WithRetryMaxAttempts(1),
	}
}

func TestAWSParamLoaderV2(t *testing.T) {
	resp, _ := json.Marshal(map[string]interface{}{
		"Parameter": map[string]interface{}{
			"Name":  "/dba/test/config",
			"Type":  "SecureString",
			"Value": testParamYAML,
		},
	})
	srv, got := fakeAWS(t, http.StatusOK, string(resp))

	p, err := cnf.AWSParamLoaderV2(context.Background(), testRegion, "/dba/test/config", testOpts(t, srv.URL)...)
	if err != nil {
		t.Fatalf("AWSParamLoaderV2 오류: %v", err)
	}

	// 요청: GetParameter 를 복호화 옵션으로, 지정한 리전 서명으로 호출해야 합니다.
	if got.Target != "AmazonSSM.GetParameter" {
		t.Errorf("X-Amz-Target = %q", got.Target)
	}
	if got.Body["Name"] != "/dba/test/config" || got.Body["WithDecryption"] != true {
		t.Errorf("요청 본문 = %v", got.Body)
	}
	if !strings.Contains(got.Authorization, "/"+testRegion+"/ssm/aws4_request") {
		t.Errorf("서명 범위에 리전이 없습니다: %q", got.Authorization)
	}

	// 응답: 기존 헬퍼가 그대로 동작해야 합니다.
	if cl := p.Conflist(); !reflect.DeepEqual(cl, []string{"repository", "server"}) {
		t.Errorf("Conflist = %v", cl)
	}
	repo := p.Keyload("repository")
	if repo["Host"] != "db.local" || repo["Port"] != 3306 {
		t.Errorf("Keyload(repository) = %v", repo)
	}
	if db := cnf.InterfaceToSlice(repo["DB"]); !reflect.DeepEqual(db, []string{"argus", "lynx"}) {
		t.Errorf("InterfaceToSlice(DB) = %v", db)
	}

	// 같은 YAML 을 FileLoader 로 읽은 결과와 동일해야 합니다(V1 도 같은 yaml.Unmarshal 경로입니다).
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(testParamYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	fromFile, err := cnf.FileLoader(path)
	if err != nil {
		t.Fatalf("FileLoader 오류: %v", err)
	}
	if !reflect.DeepEqual(p, fromFile) {
		t.Errorf("FileLoader 결과와 다릅니다\nV2  : %#v\nFile: %#v", p, fromFile)
	}
}

func TestAWSParamLoaderV2NotFound(t *testing.T) {
	srv, _ := fakeAWS(t, http.StatusBadRequest, `{"__type":"ParameterNotFound","message":"not found"}`)

	_, err := cnf.AWSParamLoaderV2(context.Background(), testRegion, "/dba/test/missing", testOpts(t, srv.URL)...)
	if err == nil {
		t.Fatal("없는 파라미터인데 오류가 없습니다")
	}
	if !strings.Contains(err.Error(), "/dba/test/missing") || !strings.Contains(err.Error(), "ParameterNotFound") {
		t.Errorf("오류에 키·원인이 없습니다: %v", err)
	}
}

func TestAWSSecretLoaderV2(t *testing.T) {
	resp, _ := json.Marshal(map[string]interface{}{
		"Name":         "dba/test/secret",
		"SecretString": `{"username":"admin","password":"p@ss","port":3306}`,
	})
	srv, got := fakeAWS(t, http.StatusOK, string(resp))

	data, err := cnf.AWSSecretLoaderV2(context.Background(), testRegion, "dba/test/secret", testOpts(t, srv.URL)...)
	if err != nil {
		t.Fatalf("AWSSecretLoaderV2 오류: %v", err)
	}

	if got.Target != "secretsmanager.GetSecretValue" {
		t.Errorf("X-Amz-Target = %q", got.Target)
	}
	if got.Body["SecretId"] != "dba/test/secret" {
		t.Errorf("요청 본문 = %v", got.Body)
	}
	if !strings.Contains(got.Authorization, "/"+testRegion+"/secretsmanager/aws4_request") {
		t.Errorf("서명 범위에 리전이 없습니다: %q", got.Authorization)
	}

	// V1 과 같이 encoding/json 으로 해석하므로 숫자는 float64 입니다.
	want := map[string]interface{}{"username": "admin", "password": "p@ss", "port": float64(3306)}
	if !reflect.DeepEqual(data, want) {
		t.Errorf("data = %v", data)
	}
}

func TestAWSSecretLoaderV2Binary(t *testing.T) {
	srv, _ := fakeAWS(t, http.StatusOK, `{"Name":"dba/test/binary","SecretBinary":"YmluYXJ5"}`)

	_, err := cnf.AWSSecretLoaderV2(context.Background(), testRegion, "dba/test/binary", testOpts(t, srv.URL)...)
	if err == nil || !strings.Contains(err.Error(), "dba/test/binary") {
		t.Fatalf("SecretBinary 는 키를 담은 오류여야 합니다: %v", err)
	}
}
