package storage

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestZZProbe(t *testing.T) {
	_ = godotenv.Load("../.env")
	s := NewR2(os.Getenv("R2_ACCOUNT_ID"), os.Getenv("R2_ACCESS_KEY_ID"), os.Getenv("R2_SECRET_ACCESS_KEY"), os.Getenv("R2_BUCKET"))
	for _, key := range []string{"location/zz-does-not-exist/none.png", "location/bandung-tengah/map.png"} {
		req, _ := http.NewRequest(http.MethodHead, s.endpoint+"/"+s.bucket+"/"+key, nil)
		s.sign(req, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", time.Now().UTC())
		resp, err := s.client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		t.Logf("HEAD %s -> %s", key, resp.Status)
	}
}
