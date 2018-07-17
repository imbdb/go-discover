package packet

import (
	"fmt"
	"log"
	"os"

	"github.com/packethost/packngo"
)

const baseURL = "https://api.packet.net/"

// Provider struct
type Provider struct {
	userAgent string
}

// SetUserAgent setter
func (p *Provider) SetUserAgent(s string) {
	p.userAgent = s
}

// Help function
func (p *Provider) Help() string {
	return `Packet:
	provider:		     "packet"
	packet_project: 	 UUID of packet project
	packet_url: 		 Packet REST URL
	packet_auth_token:   Packet authentication token
    address_type:        "private_v4", "public_v4" or "public_v6". Defaults to "private_v4".
	`
}

// Addrs function
func (p *Provider) Addrs(args map[string]string, l *log.Logger) ([]string, error) {
	authToken := argsOrEnv(args, "packet_auth_token", "PACKET_AUTH_TOKEN")
	projectID := argsOrEnv(args, "packet_project", "PACKET_PROJECT")
	packetURL := argsOrEnv(args, "packet_url", "PACKET_URL")
	addressType := args["address_type"]

	if addressType != "private_v4" && addressType != "public_v4" && addressType != "public_v6" {
		l.Printf("[INFO] discover-packet: Address type %s is not supported. Valid values are {private_v4,public_v4,public_v6}. Falling back to 'private_v4'", addressType)
		addressType = "private_v4"
	}

	c, err := client(p.userAgent, packetURL, authToken)
	if err != nil {
		return nil, fmt.Errorf("discover-packet: Initializing Packet client failed: %s", err)
	}

	var devices []packngo.Device

	if projectID == "" {
		return nil, fmt.Errorf("discover-packet: 'packet_project' parameter must be provider")
	}

	devices, _, err = c.Devices.List(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("discover-packet: Fetching Packet devices failed: %s", err)
	}
	var addrs []string
	for _, d := range devices {
		addressFamily := 4
		if addressType == "public_v6" {
			addressFamily = 6
		}
		for _, n := range d.Network {

			if (n.Public == (addressType == "public_v4" || addressType == "public_v6")) && n.AddressFamily == addressFamily {
				addrs = append(addrs, n.Address)
			}
		}
	}
	return addrs, nil
}

func client(useragent, url, token string) (*packngo.Client, error) {
	if url == "" {
		url = baseURL
	}

	return packngo.NewClientWithBaseURL(useragent, token, nil, url)
}
func argsOrEnv(args map[string]string, key, env string) string {
	if value := args[key]; value != "" {
		return value
	}
	return os.Getenv(env)
}