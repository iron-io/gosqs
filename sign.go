package sqs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"launchpad.net/goamz/aws"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func sign(auth aws.Auth, method, path string, params url.Values, headers http.Header) {
	params.Set("AWSAccessKeyId", auth.AccessKey)
	params.Set("SignatureMethod", "HmacSHA256")
	params.Set("SignatureVersion", "2")

	var sarray []string
	for k, v := range params {
		for _, vi := range v {
			sarray = append(sarray, aws.Encode(k)+"="+aws.Encode(vi))
		}
	}
	sort.StringSlice(sarray).Sort()
	joined := strings.Join(sarray, "&")

	host := headers.Get("Host")
	payload := strings.Join([]string{method, host, path, joined}, "\n")
	/*println("stringtosign")
	println(payload)
	println()*/
	hash := hmac.New(sha256.New, []byte(auth.SecretKey))
	hash.Write([]byte(payload))
	signature := make([]byte, base64.StdEncoding.EncodedLen(hash.Size()))
	base64.StdEncoding.Encode(signature, hash.Sum(nil))
	params.Set("Signature", string(signature))
}
