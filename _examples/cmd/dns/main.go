package main

import (
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/miekg/dns"

	"_examples/internal/logger"
)

var (
	log = logger.GetAppLogger()
)

// https://www.cnblogs.com/guangdelw/p/18306424
func main() {
	// 设置要查询的域名和 DNS 服务器地址
	domain := "baidu.com"
	serverAddr := "223.5.5.5:53"

	// 创建一个 DNS 消息
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	// 发送 DNS 查询
	c := new(dns.Client)
	in, _, err := c.Exchange(m, serverAddr)
	if err != nil {
		log.Errorf("发送DNS查询失败 %v", err)
		return
	}

	// 处理 DNS 响应
	if len(in.Answer) > 0 {
		for _, ans := range in.Answer {
			if a, ok := ans.(*dns.A); ok {
				log.Infof("a %v", a.A)
			}
		}
	}

	// 注册 DNS 请求处理函数
	dns.HandleFunc(".", handleDNSRequest)

	// 设置服务器地址和协议
	server := &dns.Server{Addr: ":853", Net: "udp"}
	log.Infof("Starting DNS server on %s\n", server.Addr)

	// 启动协程来扫描缓存并删除超时记录
	go cacheCleaner()
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start DNS server: %v\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Infof("Shutting down DNS server...")
	if err := server.Shutdown(); err != nil {
		log.Fatalf("Failed to shutdown DNS server: %v\n", err)
	}

}

// ARecordCacheEntry 用于存储A记录缓存信息的结构体
type ARecordCacheEntry struct {
	IP      net.IP
	TTL     uint32
	Expires time.Time
}

var (
	request = new(dns.Msg)
	// 全局变量，用于存储A记录的缓存信息
	aRecordCache = make(map[string]*ARecordCacheEntry)
	client       = &dns.Client{
		Timeout: 5 * time.Second,
	}
)

// 协程函数，用于每秒扫描缓存并删除超时记录
func cacheCleaner() {
	log.Infof("启动缓存清理协程")
	ticker := time.NewTicker(time.Second)

	for range ticker.C {
		log.Infof("starting clean cache ...")
		if len(aRecordCache) <= 0 {
			log.Infof("cache is empty")
			continue
		}
		// 遍历缓存中的所有记录
		for domain, cacheEntry := range aRecordCache {
			if time.Now().After(cacheEntry.Expires) {
				// 如果记录已过期，删除该记录
				delete(aRecordCache, domain)
				log.Infof("已删除过期的A记录缓存：%s\n", domain)
			}
		}
	}
}

// 处理 DNS 请求的函数
func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	// 将 DNS 响应标记为权威应答
	msg.Authoritative = true
	// 将 DNS 响应标记为递归可用
	// msg.RecursionAvailable = true

	// 遍历请求中的问题部分，生成相应的回答
	for _, question := range r.Question {
		log.Infof("请求解析的域名：", question.Name)
		switch question.Qtype {
		case dns.TypeA:
			log.Infof("Query Type A")
			handleARecord(question, msg)
		case dns.TypeAAAA:
			log.Infof("Query Type AAAA")
			handleAAAARecord(question, msg)
		// 你可以在这里添加其他类型的记录处理逻辑
		case dns.TypeMX:
			log.Infof("Query Type MX")
			handleMXRecord(question, msg)
		case dns.TypeTXT:
			log.Infof("Query Type TXT")
			handleTXTRecord(question, msg)
		case dns.TypeCNAME:
			log.Infof("Query Type CNAME")
			handleCNAMERecord(question, msg)
		case dns.TypeNS:
			log.Infof("Query Type NS")
			handleNSRecord(question, msg)
		}
	}
	if err := w.WriteMsg(msg); err != nil {
		log.Infof("Failed to write DNS response:", err)
	}
}

// 构建 A 记录的函数
func handleARecord(q dns.Question, msg *dns.Msg) {
	domain := q.Name
	// 先检查缓存中是否存在该域名的记录
	if cacheEntry, ok := aRecordCache[domain]; ok {
		// 判断TTL是否超时
		if time.Now().Before(cacheEntry.Expires) {
			aRecord := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    cacheEntry.TTL,
				},
				A: cacheEntry.IP,
			}
			msg.Answer = append(msg.Answer, aRecord)
			log.Infof("从缓存中获取A记录： ", cacheEntry.IP.String(), " ttl ", cacheEntry.TTL)
			return
		} else {
			// TTL超时，从缓存中删除该记录
			delete(aRecordCache, domain)
		}
	}

	request.SetQuestion(dns.Fqdn(q.Name), dns.TypeA)
	response, rtt, err := client.Exchange(request, "223.5.5.5:53")
	if err != nil {
		log.Infof("Failed to resolve A record for %s: %v\n", q.Name, err)
		return
	}
	log.Infof("Resolved A record for %s in %v\n", q.Name, rtt)
	if response.Rcode != dns.RcodeSuccess {
		log.Infof("Failed to resolve A record for %s: %v\n", q.Name, response.Rcode)
		return
	}
	// 判断是否来自权威服务器
	if response.Authoritative {
		log.Infof("来自权威服务器")
	}
	// 判断是否递归查询
	if response.RecursionAvailable {
		log.Infof("这是递归查询")
	}
	for _, rr := range response.Answer {
		if a, ok := rr.(*dns.A); ok {
			log.Infof("解析结果：", a.A.String(), " ttl ", a.Hdr.Ttl)

			// 将新查询到的结果存入缓存
			cacheEntry := &ARecordCacheEntry{
				IP:      a.A,
				TTL:     a.Hdr.Ttl,
				Expires: time.Now().Add(time.Duration(a.Hdr.Ttl) * time.Second),
			}
			aRecordCache[domain] = cacheEntry

			msg.Answer = append(msg.Answer, a)
		}
	}
	// ip := net.ParseIP("192.0.2.1")
	// rr := &dns.A{
	//     Hdr: dns.RR_Header{
	//         Name:   q.Name,
	//         Rrtype: dns.TypeA,
	//         Class:  dns.ClassINET,
	//         Ttl:    600,
	//     },
	//     A: ip,
	// }
	// msg.Answer = append(msg.Answer, rr)
}

func handleAAAARecord(q dns.Question, msg *dns.Msg) {
	request.SetQuestion(dns.Fqdn(q.Name), dns.TypeAAAA)
	response, rtt, err := client.Exchange(request, "223.5.5.5:53")
	if err != nil {
		log.Infof("Failed to resolve AAAA record for %s: %v\n", q.Name, err)
		return
	}

	log.Infof("Resolved AAAA record for %s in %v\n", q.Name, rtt)
	if response.Rcode != dns.RcodeSuccess {
		log.Infof("Failed to resolve AAAA record for %s: %v\n", q.Name, response.Rcode)
		return
	}
	// 判断是否来自权威服务器
	if response.Authoritative {
		log.Infof("来自权威服务器")
	}
	// 判断是否递归查询
	if response.RecursionAvailable {
		log.Infof("这是递归查询")
	}
	for _, rr := range response.Answer {
		if aaaa, ok := rr.(*dns.AAAA); ok {
			msg.Answer = append(msg.Answer, aaaa)
		}
	}
	// ip := net.ParseIP("240c::6666")
	// rr := &dns.AAAA{
	//     Hdr: dns.RR_Header{
	//         Name:   q.Name,
	//         Rrtype: dns.TypeAAAA,
	//         Class:  dns.ClassINET,
	//         Ttl:    600,
	//     },
	//     AAAA: ip,
	// }
	// msg.Answer = append(msg.Answer, rr)
}

func handleCNAMERecord(q dns.Question, msg *dns.Msg) {
	request.SetQuestion(dns.Fqdn(q.Name), dns.TypeCNAME)
	response, rtt, err := client.Exchange(request, "223.5.5.5:53")
	if err != nil {
		log.Infof("Failed to resolve CNAME record for %s: %v\n", q.Name, err)
		return
	}
	log.Infof("Resolved CNAME record for %s in %v\n", q.Name, rtt)
	if response.Rcode != dns.RcodeSuccess {
		log.Infof("Failed to resolve CNAME record for %s: %v\n", q.Name, response.Rcode)
		return
	}
	// 判断是否来自权威服务器
	if response.Authoritative {
		log.Infof("来自权威服务器")
	}
	// 判断是否递归查询
	if response.RecursionAvailable {
		log.Infof("这是递归查询")
	}
	for _, rr := range response.Answer {
		if cname, ok := rr.(*dns.CNAME); ok {
			log.Infof("CNAME :", cname)
			msg.Answer = append(msg.Answer, cname)
		}
	}
	// rr := &dns.CNAME{
	//     Hdr: dns.RR_Header{
	//         Name:   q.Name,
	//         Rrtype: dns.TypeCNAME,
	//         Class:  dns.ClassINET,
	//         Ttl:    600,
	//     },
	//     Target: "example.com.",
	// }
	// msg.Answer = append(msg.Answer, rr)
}

func handleMXRecord(q dns.Question, msg *dns.Msg) {

	request.SetQuestion(dns.Fqdn(q.Name), dns.TypeMX)
	response, rtt, err := client.Exchange(request, "223.5.5.5:53")
	if err != nil {
		log.Infof("Failed to resolve MX record for %s: %v\n", q.Name, err)
		return
	}
	log.Infof("Resolved MX record for %s in %v\n", q.Name, rtt)
	if response.Rcode != dns.RcodeSuccess {
		log.Infof("Failed to resolve MX record for %s: %v\n", q.Name, response.Rcode)
		return
	}
	// 判断是否来自权威服务器
	if response.Authoritative {
		log.Infof("来自权威服务器")
	}
	// 判断是否递归查询
	if response.RecursionAvailable {
		log.Infof("这是递归查询")
	}
	for _, rr := range response.Answer {
		if mx, ok := rr.(*dns.MX); ok {
			log.Infof("MX :", mx)
			msg.Answer = append(msg.Answer, mx)

		}
	}
	// rr := &dns.MX{
	//     Hdr: dns.RR_Header{
	//         Name:   q.Name,
	//         Rrtype: dns.TypeMX,
	//         Class:  dns.ClassINET,
	//         Ttl:    600,
	//     },
	//     Preference: 10,
	//     Mx:         "mail.example.com.",
	// }
	// msg.Answer = append(msg.Answer, rr)
}

func handleTXTRecord(q dns.Question, msg *dns.Msg) {
	request.SetQuestion(dns.Fqdn(q.Name), dns.TypeTXT)
	response, rtt, err := client.Exchange(request, "223.5.5.5:53")
	if err != nil {
		log.Infof("Failed to resolve TXT record for %s: %v\n", q.Name, err)
		return
	}
	log.Infof("Resolved TXT record for %s in %v\n", q.Name, rtt)
	if response.Rcode != dns.RcodeSuccess {
		log.Infof("Failed to resolve TXT record for %s: %v\n", q.Name, response.Rcode)
		return
	}
	// 判断是否来自权威服务器
	if response.Authoritative {
		log.Infof("来自权威服务器")
	}
	// 判断是否递归查询
	if response.RecursionAvailable {
		log.Infof("这是递归查询")
	}
	for _, rr := range response.Answer {
		if txt, ok := rr.(*dns.TXT); ok {
			log.Infof("TXT:", txt.Txt)
			msg.Answer = append(msg.Answer, txt)
		}
	}

	// rr := &dns.TXT{
	//     Hdr: dns.RR_Header{
	//         Name:   q.Name,
	//         Rrtype: dns.TypeTXT,
	//         Class:  dns.ClassINET,
	//         Ttl:    600,
	//     },
	//     Txt: []string{"v=spf1 include:_spf.example.com ~all"},
	// }
	// msg.Answer = append(msg.Answer, rr)
}

func handleNSRecord(q dns.Question, msg *dns.Msg) {

	request.SetQuestion(dns.Fqdn(q.Name), dns.TypeNS)

	response, rtt, err := client.Exchange(request, "223.5.5.5:53")
	if err != nil {
		log.Infof("Failed to resolve NS record for %s: %v\n", q.Name, err)
		return
	}
	log.Infof("Resolved NS record for %s in %v\n", q.Name, rtt)
	if response.Rcode != dns.RcodeSuccess {
		log.Infof("Failed to resolve NS record for %s: %v\n", q.Name, response.Rcode)
		return
	}
	// 判断是否来自权威服务器
	if response.Authoritative {
		log.Infof("来自权威服务器")
	}
	// 判断是否递归查询
	if response.RecursionAvailable {
		log.Infof("这是递归查询")
	}
	for _, rr := range response.Answer {
		if ns, ok := rr.(*dns.NS); ok {
			log.Infof("NS:", ns.Ns)
			msg.Answer = append(msg.Answer, ns)
		}
	}
}
