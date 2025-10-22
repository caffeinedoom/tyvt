# TLS Certificate Verification with Proxies - Complete Guide

**Issue**: `tls: failed to verify certificate: x509: certificate signed by unknown authority`

This guide explains why this happens with residential proxies (like Bright Data) and provides multiple solutions with security considerations.

---

## 🔍 **Understanding the Problem**

### What's Happening?

When you use a proxy like Bright Data with HTTPS endpoints:

```
Your App ─→ HTTP Proxy (Bright Data) ─→ HTTPS (VirusTotal)
              ↓
        TLS Inspection
              ↓
    Proxy's Certificate ✗ (Unknown CA)
```

**Normal HTTPS Flow (No Proxy):**
```
Your App ─────→ VirusTotal
  ↓ TLS Handshake
  ↓ Verifies VirusTotal's Certificate ✓
  ↓ Encrypted Connection
```

**With Proxy Performing TLS Inspection:**
```
Your App ─→ Proxy ─→ VirusTotal
  ↓           ↓
  ↓       Proxy intercepts
  ↓       Presents own cert
  ↓
  ✗ Go rejects (unknown CA)
```

### Why Do Proxies Do This?

Residential proxy providers like Bright Data perform **TLS inspection** for:

1. **Traffic Analysis**: Monitor for abuse/violations
2. **Caching**: Improve performance
3. **Compliance**: Ensure terms of service adherence
4. **Security**: Detect malicious traffic
5. **Billing**: Track bandwidth usage accurately

This is common with:
- Bright Data (BRD)
- Oxylabs
- Smartproxy
- Luminati (now Bright Data)

---

## ✅ **Solution 1: Skip TLS Verification** (Quick Fix)

### Usage:

```bash
# Add the --insecure-tls flag
./tyvt -d domains.txt -k keys.txt \
  -p 'http://user:pass@brd.superproxy.io:33335' \
  --insecure-tls
```

### What Happens:

```go
// Go skips certificate validation
transport.TLSClientConfig = &tls.Config{
    InsecureSkipVerify: true,
}
```

### Security Warning:

When you run this, you'll see:

```
⚠️  TLS certificate verification is DISABLED
    This makes you vulnerable to man-in-the-middle attacks
    Only use this with trusted proxy providers
```

### When to Use:

✅ **Safe to use when:**
- You trust your proxy provider (Bright Data, Oxylabs, etc.)
- You're testing functionality
- Proxy provider confirms they perform TLS inspection
- You're on a secure network

❌ **DO NOT use when:**
- Using untrusted/free proxies
- On public WiFi
- Handling sensitive data
- Compliance requires certificate validation

---

## ✅ **Solution 2: Add Proxy CA Certificate** (Recommended)

This is the **most secure** approach - import the proxy provider's root CA certificate.

### Step 1: Get the CA Certificate

**Option A: From Bright Data Support**
```bash
# Contact Bright Data support and request their root CA certificate
# They should provide a .crt or .pem file
```

**Option B: Extract from Connection** (Advanced)
```bash
# Connect through proxy and extract certificate
openssl s_client -connect virustotal.com:443 \
  -proxy brd.superproxy.io:33335 \
  -showcerts
```

### Step 2: System-Wide Install (Linux)

```bash
# Copy certificate to system trust store
sudo cp brightdata-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Restart your application
./tyvt -d domains.txt -k keys.txt -p 'http://user:pass@proxy.com:8080'
# No --insecure-tls needed!
```

### Step 3: Application-Specific (Go Code)

If you can't modify system trust store, we can add it to the application:

```go
// Load CA certificate
caCert, err := ioutil.ReadFile("brightdata-ca.crt")
if err != nil {
    log.Fatal(err)
}

// Create cert pool
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

// Configure transport
transport.TLSClientConfig = &tls.Config{
    RootCAs: caCertPool,
}
```

### Advantages:
- ✅ Maintains security
- ✅ Validates certificates properly
- ✅ No security warnings
- ✅ Production-ready

---

## ✅ **Solution 3: Contact Bright Data Support**

### What to Ask:

```
Subject: TLS Certificate Verification Error with Go Client

Hi Bright Data Support,

I'm using your residential proxy service with a Go application and 
receiving this error:

"tls: failed to verify certificate: x509: certificate signed by unknown authority"

Questions:
1. Can you provide your root CA certificate for installation?
2. Can TLS inspection be disabled for my account?
3. What's the recommended configuration for Go HTTP clients?

Zone: residential_proxy2
Account: hl_b0fb7729
```

### Possible Outcomes:

1. **They provide CA certificate** → Use Solution 2
2. **They disable TLS inspection** → Problem solved automatically
3. **They recommend insecure mode** → Use Solution 1 with confidence

---

## 🔧 **Diagnosing the Issue**

### Test 1: Verify Proxy Works Without TLS

```bash
# Test HTTP endpoint (no TLS)
curl -x http://user:pass@brd.superproxy.io:33335 http://example.com

# If this works, proxy is functional
# If this fails, check proxy credentials/connectivity
```

### Test 2: Check Certificate Chain

```bash
# See what certificate the proxy presents
openssl s_client -connect virustotal.com:443 \
  -proxy brd.superproxy.io:33335 \
  -showcerts 2>&1 | grep -A 10 "Certificate chain"
```

### Test 3: Test Direct Connection

```bash
# Run TYVT without proxy
./tyvt -d domains.txt -k keys.txt -o results.txt

# If this works, issue is definitely proxy-related
```

---

## 📊 **Comparison of Solutions**

| Solution | Security | Ease | Production | Setup Time |
|----------|----------|------|------------|------------|
| **--insecure-tls** | ⚠️ Low | ✅ Easy | ❌ Not recommended | 1 min |
| **CA Certificate** | ✅ High | ⚠️ Medium | ✅ Recommended | 30 min |
| **Contact Support** | ✅ High | ⚠️ Medium | ✅ Best | 1-2 days |

---

## 🚀 **Quick Start (Testing)**

If you need to test **right now**:

```bash
# Build with new flag
cd /root/pluckware/tyvt
go build -o tyvt .

# Run with TLS verification disabled
./tyvt \
  -d /root/wonder/programs/rikhter/tryouts/urls/to_query.txt \
  -k keys.txt \
  -o /root/wonder/programs/rikhter/tryouts/urls/tyvt.txt \
  -p 'http://brd-customer-hl_b0fb7729-zone-residential_proxy2:x03yd8oxs3s6@brd.superproxy.io:33335' \
  --insecure-tls
```

**Expected Output:**
```
✓ Using proxy: http://brd.superproxy.io:33335

[2025-10-22 23:15:00] WARN: ⚠️  TLS certificate verification is DISABLED
[2025-10-22 23:15:00] WARN:     This makes you vulnerable to man-in-the-middle attacks
[2025-10-22 23:15:00] WARN:     Only use this with trusted proxy providers
[2025-10-22 23:15:00] INFO: Starting scan of 5 domains with 2 API keys
[2025-10-22 23:15:00] INFO: Scanning domain 1/5: example.com
...
```

---

## 🔐 **Security Best Practices**

### 1. For Testing/Development

```bash
# Use --insecure-tls (acceptable risk)
./tyvt -d domains.txt -k keys.txt -p $PROXY --insecure-tls
```

### 2. For Production

```bash
# Option A: Install CA certificate system-wide
sudo cp proxy-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
./tyvt -d domains.txt -k keys.txt -p $PROXY

# Option B: Request TLS inspection be disabled
# (Contact proxy provider)
```

### 3. Environment Variables

```bash
# Store proxy URL securely
export TYVT_PROXY='http://user:pass@proxy.com:8080'

# Use in script
./tyvt -d domains.txt -k keys.txt -p "$TYVT_PROXY" --insecure-tls
```

---

## 🐛 **Troubleshooting**

### Error: Still Getting Certificate Error

```bash
# Verify flag is being recognized
./tyvt -h | grep insecure

# Check you're using the right flag format
./tyvt --insecure-tls ...  # Correct
./tyvt -insecure-tls ...   # Also works
./tyvt insecure-tls ...    # Wrong (missing dashes)
```

### Error: Proxy Authentication Failed (407)

```bash
# This is DIFFERENT from TLS error
# Verify credentials
curl -x 'http://user:pass@proxy.com:8080' https://google.com

# URL-encode special characters
# @ → %40
# : → %3A
# ! → %21
```

### Connection Times Out

```bash
# Test proxy connectivity
nc -zv brd.superproxy.io 33335

# If this fails, check:
# 1. Firewall rules
# 2. Proxy IP whitelist (some providers require it)
# 3. Network connectivity
```

---

## 📚 **Additional Resources**

### Bright Data Specific

- **Documentation**: https://docs.brightdata.com/
- **Support**: support@brightdata.com
- **Status Page**: https://status.brightdata.com/

### Go TLS Configuration

- [crypto/tls Package](https://pkg.go.dev/crypto/tls)
- [x509 Package](https://pkg.go.dev/crypto/x509)
- [Transport TLS Config](https://pkg.go.dev/net/http#Transport)

### Certificate Management

- [Ubuntu CA Certificates](https://ubuntu.com/server/docs/security-trust-store)
- [OpenSSL s_client](https://www.openssl.org/docs/man1.1.1/man1/s_client.html)

---

## 🎯 **Recommendations**

### For Your Use Case (Bright Data):

**Immediate (Testing):**
```bash
./tyvt -d domains.txt -k keys.txt -p $PROXY --insecure-tls
```
✅ Gets you running immediately  
⚠️ Understand security trade-offs  
📧 Contact Bright Data for long-term solution

**Long-term (Production):**
1. Contact Bright Data support
2. Request root CA certificate OR disable TLS inspection
3. Implement proper solution (CA cert or no inspection)
4. Remove `--insecure-tls` flag

---

## ⚡ **Summary**

**Your Error:**
```
tls: failed to verify certificate: x509: certificate signed by unknown authority
```

**Root Cause:**
Bright Data (like many residential proxy providers) performs TLS inspection, presenting their own certificate that Go doesn't trust.

**Quick Fix:**
```bash
./tyvt ... -p $PROXY --insecure-tls
```

**Proper Fix:**
1. Contact Bright Data for CA certificate
2. Install CA cert to system trust store
3. Run without `--insecure-tls` flag

**Security Note:**
Using `--insecure-tls` is acceptable with **trusted** proxy providers like Bright Data. Just understand you're trusting them completely with your traffic.

---

## 🎉 **Ready to Use**

Your command should now work:

```bash
./tyvt \
  -d /root/wonder/programs/rikhter/tryouts/urls/to_query.txt \
  -k keys.txt \
  -o /root/wonder/programs/rikhter/tryouts/urls/tyvt.txt \
  -p 'http://brd-customer-hl_b0fb7729-zone-residential_proxy2:x03yd8oxs3s6@brd.superproxy.io:33335' \
  --insecure-tls
```

The TLS error will be gone! 🎊

**Questions?** This is a common issue with residential proxies. You're doing nothing wrong - it's just how these services operate.
