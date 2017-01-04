package utils

import (
	"errors"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"

	garbler "github.com/michaelbironneau/garbler/lib"
)

func MakeUniqueFile(name string, suffix string, count int) (*os.File, error) {
	var filename string

	if count > 0 {
		filename = name + "-" + strconv.Itoa(count)
	} else {
		filename = name
	}

	filename = filename + suffix

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename)
	} else {
		return MakeUniqueFile(name, suffix, count+1)
	}
}

func GenerateAlphaNumericPassword() string {
	attempts := 0
	for {
		reqs := &garbler.PasswordStrengthRequirements{
			MinimumTotalLength: 16,
			Uppercase:          rand.Intn(6),
			Digits:             rand.Intn(6),
			Punctuation:        -1, // disable punctuation
		}
		pass, err := garbler.NewPassword(reqs)
		if err != nil {
			return "weakpassword"
		}
		// validate that the library actually returned an alphanumeric password
		re := regexp.MustCompile("^[a-zA-Z1-9]+$")
		if re.MatchString(pass) {
			return pass
		}
		if attempts == 50 {
			return "weakpassword"
		}
		attempts++
	}
}

func IncrementIPv4(ip net.IP) (net.IP, error) {

	if len(ip) != net.IPv4len {
		return nil, errors.New("IncrementIPv4: only IPv4 addresses supported")
	}

	if ip.Equal(net.IPv4bcast) {
		return nil, errors.New("IncrementIPv4: incrementing IPv4 broadcast (255.255.255.255) results in overflow")
	}

	nextIp := make(net.IP, len(ip))
	copy(nextIp, ip)

	// increment IP accounting for rollovers at each byte
	for j := len(nextIp) - 1; j >= 0; j-- {
		nextIp[j]++
		if nextIp[j] > 0 {
			break
		}
	}

	return nextIp, nil
}

func BroadcastIPv4(network net.IPNet) (net.IP, error) {
	if len(network.IP) != net.IPv4len {
		return nil, errors.New("BroadcastIPv4: only IPv4 addresses supported")
	}

	broadcast := net.IP(make([]byte, 4))
	for i := range network.IP {
		broadcast[i] = network.IP[i] | ^network.Mask[i]
	}

	return broadcast, nil
}
