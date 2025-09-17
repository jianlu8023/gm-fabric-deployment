package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"net/http"

	"github.com/miekg/dns"
)

func handleDNSQuery(w http.ResponseWriter, r *http.Request) {
	log.Println("received request")
	var question []byte
	var err error
	if r.Method == http.MethodGet {
		q := r.URL.Query().Get("dns")
		question, err = base64.RawURLEncoding.DecodeString(q)
	} else {
		question, err = io.ReadAll(r.Body)
		r.Body.Close()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var m dns.Msg
	if err := m.Unpack(question); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// var hasSubnet bool
	// if e := m.IsEdns0(); e != nil {
	//     for _, o := range e.Option {
	//         if o.Option() == dns.EDNS0SUBNET {
	//             a := o.(*dns.EDNS0_SUBNET).Address[:2]
	//             // skip empty subnet like 0.0.0.0/0
	//             if !bytes.HasPrefix(a, []byte{0, 0}) {
	//                 hasSubnet = true
	//             }
	//             break
	//         }
	//     }
	// }

	// if !hasSubnet {
	//     ip, err := netip.ParseAddrPort(r.RemoteAddr)
	//     if err != nil {
	//         http.Error(w, err.Error(), http.StatusInternalServerError)
	//         return
	//     }
	//     addr := ip.Addr()
	//     opt := &dns.OPT{Hdr: dns.RR_Header{Name: ".", Rrtype: dns.TypeOPT}}
	//     ecs := &dns.EDNS0_SUBNET{Code: dns.EDNS0SUBNET}
	//     var bits int
	//     if addr.Is4() {
	//         bits = 24
	//         ecs.Family = 1
	//     } else {
	//         bits = 48
	//         ecs.Family = 2
	//     }
	//     ecs.SourceNetmask = uint8(bits)
	//     p := netip.PrefixFrom(addr, bits)
	//     ecs.Address = net.IP(p.Masked().Addr().AsSlice())
	//     opt.Option = append(opt.Option, ecs)
	//     m.Extra = []dns.RR{opt}
	// }

	if question, err = m.Pack(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := http.Post("https://doh.pub/dns-query", "application/dns-message", bytes.NewReader(question))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	answer, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("content-type", "application/dns-message")
	w.Write(answer)

}

// https://taoshu.in/dns/diy-doh.html
func main() {
	http.HandleFunc("/dns-query", handleDNSQuery)

	log.Println("Listening on :8053")
	log.Fatal(http.ListenAndServeTLS(":8053", "./cert/rsa.crt", "./cert/rsa.key", nil))
}
