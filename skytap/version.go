package skytap

const userAgentVersion = `{
  "version": "0.11.x"
}`

type userAgent struct {
	Version string `json:"version"`
}
