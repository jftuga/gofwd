package main

func getExamples() [][]string {
	examples := [][]string{}

	examples = append(examples, []string{`get the local IP address *(run this first)*, eg: 1.2.3.4`, `gofwd -i `})
	examples = append(examples, []string{`forward from a bastion host to an internal server`, `gofwd -f 1.2.3.4:22 -t 192.168.1.1:22`})
	examples = append(examples, []string{`allow only if the remote IP is within 50 miles of this host`, `gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 -d 50`})
	examples = append(examples, []string{`allow only if remote IP is located in Denver, CO`, `gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 -city Denver -region Colorado`})
	examples = append(examples, []string{`allow only if remote IP is located in Canada`, `gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 -country CA`})
	examples = append(examples, []string{`allow only if remote IP is located within 75 miles of Atlanta, GA`, `gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 -l 33.756529,-84.400996 -d 75`})
	examples = append(examples, []string{`    to get Latitude, Longitude use https://www.latlong.net/`, ` `})
	examples = append(examples, []string{`allow only for a successful two-factor duo auth for 'testuser'`, `gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 --duo duo.ini:testuser`})
	examples = append(examples, []string{`allow only after both Geo IP and Duo are verified`, `gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 --region Texas --duo duo.ini:testuser`})
	examples = append(examples, []string{`forward from any interface on port 22, allow RFC1918 to connect`, `gofwd -f 0.0.0.0:22 -t 192.168.1.1:22 -p`})
	examples = append(examples, []string{`forward from IP address bounded to eth0, allow RFC1918 to connect`, `gofwd -f _eth0:22 -t 192.168.1.1:22 -p`})
	examples = append(examples, []string{`forward from IP address bounded to eno1, allow RFC1918 to connect`, `gofwd -f _eno1:80 -t example.com:80 -p`})

	return examples
}
