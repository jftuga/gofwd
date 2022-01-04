/*
Adopted from:
https://github.com/jftuga/nics
*/

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func isBriefEntry(ifaceName, macAddr, mtu, flags string, ipv4List, ipv6List []string, debug bool) bool {
	if debug {
		fmt.Println("isBriefEntry:", ifaceName)
	}
	if strings.Contains(flags, "loopback") {
		if debug {
			fmt.Println("   not_brief: loopback flag")
		}
		return false
	}
	if strings.HasPrefix(macAddr, "00:00:00:00:00:00") {
		if debug {
			fmt.Println("   not_brief: NULL macAddr")
		}
		return false
	}
	if 0 == len(ipv4List) {
		if debug {
			fmt.Println("   not_brief: no IP addresses")
		}
		return false
	}
	for _, ipv4 := range ipv4List {
		if strings.HasPrefix(ipv4, "169.254.") {
			if debug {
				fmt.Println("   not_brief: self assigned:", ipv4)
			}
			return false
		}
	}
	if debug {
		fmt.Println("    is_brief: true")
	}
	return true
}

func extractIPAddrs(ifaceName string, allAddresses []net.Addr, brief bool) ([]string, []string) {
	var allIPv4 []string
	var allIPv6 []string

	for _, netAddr := range allAddresses {
		address := netAddr.String()
		if strings.Contains(address, ":") {
			allIPv6 = append(allIPv6, address)
		} else {
			allIPv4 = append(allIPv4, address)
		}
	}
	return allIPv4, allIPv6
}

func networkInterfaces(brief bool, debug bool) ([]string, []string, error) {
	adapters, err := net.Interfaces()
	if err != nil {
		return nil, nil, err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	if brief {
		table.SetHeader([]string{"Name", "IPv4", "Mac Address", "MTU", "Flags"})
	} else {
		table.SetHeader([]string{"Name", "IPv4", "IPv6", "Mac Address", "MTU", "Flags"})
	}

	var v4Addresses []string
	var v6Addresses []string
	for _, iface := range adapters {
		allAddresses, err := iface.Addrs()
		if err != nil {
			return nil, nil, nil
		}

		allIPv4, allIPv6 := extractIPAddrs(iface.Name, allAddresses, brief)
		if debug {
			fmt.Println()
			fmt.Println("---------------------")
			fmt.Println(iface.Name, allAddresses)
			fmt.Println("ipv4:", allIPv4)
			fmt.Println("ipv6:", allIPv6)
		}

		ifaceName := strings.ToLower(iface.Name)
		macAddr := iface.HardwareAddr.String()
		mtu := strconv.Itoa(iface.MTU)
		flags := iface.Flags.String()

		if brief && isBriefEntry(ifaceName, macAddr, mtu, flags, allIPv4, allIPv6, debug) {
			table.Append([]string{iface.Name, strings.Join(allIPv4, "\n"), macAddr, mtu, flags})
			for _, ipWithMask := range allIPv4 {
				ip := strings.Split(ipWithMask, "/")
				v4Addresses = append(v4Addresses, ip[0])
			}
			continue
		}

		if !brief {
			table.SetAutoWrapText(true)
			table.SetRowLine(true)
			table.Append([]string{ifaceName, strings.Join(allIPv4, "\n"), strings.Join(allIPv6, "\n"), macAddr, mtu, strings.Replace(flags, "|", "\n", -1)})
			for _, ipWithMask := range allIPv4 {
				ip := strings.Split(ipWithMask, "/")
				v4Addresses = append(v4Addresses, ip[0])
			}
			for _, ipWithMask := range allIPv6 {
				ip := strings.Split(ipWithMask, "/")
				v6Addresses = append(v6Addresses, ip[0])
			}
		}
	}
	table.Render()

	return v4Addresses, v6Addresses, err
}

func nics() {
	argsAllDetails := false
	argsDebug := false
	_, _, err := networkInterfaces(!(argsAllDetails), argsDebug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
