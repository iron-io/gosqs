package sqs

import (
	"crypto/hmac"
	"encoding/base64"
	"http"
	"launchpad.net/goamz/aws"
	"sort"
	"strings"
	"url"
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
	hash := hmac.NewSHA256([]byte(auth.SecretKey))
	hash.Write([]byte(payload))
	signature := make([]byte, base64.StdEncoding.EncodedLen(hash.Size()))
	base64.StdEncoding.Encode(signature, hash.Sum())
	params.Set("Signature", string(signature))
}
