```python
readme_en_content = """# 🚀 LiteObfsVPN

[![License: Unlicense / Public Domain](https://img.shields.io/badge/License-Unlicense-blue.svg)](https://unlicense.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows-blue)](https://github.com/)
[![Status](https://img.shields.io/badge/status-stable-green)](https://github.com/)

**LiteObfsVPN** is an ultra-lightweight, secure, and minimalistic single-binary VPN server and client tailored for personal use. The project focuses on high performance, reliable encryption, and effective traffic obfuscation to seamlessly bypass Deep Packet Inspection (DPI) systems.

---

## ✨ Key Features

* **Single Binary Concept:** The exact same compiled codebase acts as both the server and the client. The operation mode is automatically detected based on the host OS:
    * 🐧 **Linux** — automatically launches in **Server Mode**.
    * 🪟 **Windows** — automatically launches in **Client Mode**.
* **Obfuscation & Encryption:** All tunneled traffic is transformed into a high-entropy pseudorandom binary stream and end-to-end encrypted, making it fully immune to protocol-based DPI blocks.
* **Smart Autodetection:** Zero-configuration automatic detection of the external WAN interface for NAT on the server side, and the physical default gateway on the client side.
* **High Performance:** Achieves bare-metal speeds and minimal overhead by leveraging native **TUN/TAP** interfaces on Linux and the robust **Wintun** driver on Windows.

---

## 🛠️ Architecture & CLI Flags

### 🐧 Server Mode (Linux)

When executed on a Linux environment, the binary acts as a full-featured VPN gateway, automatically configuring IP forwarding and NAT rules.

```bash
sudo ./liteobfsvpn [flags]

```

| Flag | Description | Default Value |
| --- | --- | --- |
| `-listen` | IP address and port to bind for incoming client connections | `0.0.0.0:443` |
| `-egress` | External network interface used for NAT routing | *Autodetected* |
| `-key` | Unique shared secret/cryptographic key for authorization | *Built-in default key* |

*Example usage:*

```bash
sudo ./liteobfsvpn -listen "0.0.0.0:8443" -key "my_super_secret_crypto_key_2026"

```

---

### 🪟 Client Mode (Windows)

When executed on Windows, the binary provisions a virtual network adapter and securely routes all global system traffic through the obfuscated tunnel.

> ⚠️ **Important:** To run the client successfully, the **`wintun.dll`** file must reside in the exact same directory as your executable (`liteobfsvpn.exe`).

```cmd
liteobfsvpn.exe [flags]

```

| Flag | Description | Default Value |
| --- | --- | --- |
| `-server` | Remote IP address and port of your VPN server | *Required parameter* |
| `-gw` | Physical gateway IP address (router local IP) | *Autodetected* |
| `-dns` | Custom DNS server to provision for the tunnel interface | `1.1.1.1` (Cloudflare) |
| `-key` | Unique shared cryptographic key matching the server | *Built-in default key* |

*Example usage (Run as Administrator):*

```cmd
liteobfsvpn.exe -server "203.0.113.50:8443" -key "my_super_secret_crypto_key_2026" -dns "8.8.8.8"

```

---

## 🚀 Quick Start

### Step 1: Deploy Server (Linux)

1. Transfer or build the binary on your remote Linux VPS.
2. Fire up the listener:
```bash
sudo ./liteobfsvpn -listen "0.0.0.0:443" -key "your_personal_key"

```



### Step 2: Configure Client (Windows)

1. Move the compiled `liteobfsvpn.exe` to your local machine.
2. Download the official `wintun.dll` matching your system architecture (x64/x86) and drop it next to `liteobfsvpn.exe`.
3. Open a Command Prompt (CMD) or PowerShell session **as Administrator** and execute:
```cmd
liteobfsvpn.exe -server "YOUR_SERVER_IP:443" -key "your_personal_key"

```


4. Verify your secure routing via any public IP checking service.

---

## 🔒 Security & Traffic Stealthing

1. **High-Entropy Noise:** The obfuscation layer strips away standard protocol indicators. To any middleboxes or ISPs, the connection looks like an unclassifiable stream of random bytes, eliminating DPI fingerprinting signatures (unlike OpenVPN or WireGuard).
2. **Key Security:** If the `-key` flag is omitted, a hardcoded fallback token is utilized. For robust security, **always** enforce a custom, high-entropy key string.

---

## 🛠️ Compilation (Development)

The project natively supports seamless cross-compilation out of the box:

```bash
# Compile for Linux Server
GOOS=linux GOARCH=amd64 go build -o liteobfsvpn main.go

# Compile for Windows Client
GOOS=windows GOARCH=amd64 go build -o liteobfsvpn.exe main.go

```

---

## 📄 License & Disclaimer

### "AS IS" Disclaimer

This software is provided **"AS IS"**, without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no event shall the authors or copyright holders be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the software or the use or other dealings in the software.

### Unlicense / Public Domain

This is free and unencumbered software released into the public domain. Anyone is free to copy, modify, publish, use, compile, sell, or distribute this software, either in source code form or as a compiled binary, for any purpose, commercial or non-commercial, and by any means. For more information, please refer to the [Unlicense](https://unlicense.org/) details.
"""

with open("README_EN.md", "w", encoding="utf-8") as f:
f.write(readme_en_content)

print("README_EN.md successfully created!")

```

```text?code_stdout&code_event_index=1
README_EN.md successfully created!


```

Your Markdown file is ready


Я перевёл файл на английский язык, обновил лицензионные бэджи и добавил строгий дисклеймер **"AS IS"** (как есть) наряду со свободной лицензией (Unlicense / Public Domain), полностью освобождающей код от ограничений.

Вы можете использовать этот текст для файла `README.md` вашего англоязычного репозитория:

```markdown
# 🚀 LiteObfsVPN

[![License: Unlicense / Public Domain](https://img.shields.io/badge/License-Unlicense-blue.svg)](https://unlicense.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows-blue)](https://github.com/)
[![Status](https://img.shields.io/badge/status-stable-green)](https://github.com/)

**LiteObfsVPN** is an ultra-lightweight, secure, and minimalistic single-binary VPN server and client tailored for personal use. The project focuses on high performance, reliable encryption, and effective traffic obfuscation to seamlessly bypass Deep Packet Inspection (DPI) systems.

---

## ✨ Key Features

* **Single Binary Concept:** The exact same compiled codebase acts as both the server and the client. The operation mode is automatically detected based on the host OS:
    * 🐧 **Linux** — automatically launches in **Server Mode**.
    * 🪟 **Windows** — automatically launches in **Client Mode**.
* **Obfuscation & Encryption:** All tunneled traffic is transformed into a high-entropy pseudorandom binary stream and end-to-end encrypted, making it fully immune to protocol-based DPI blocks.
* **Smart Autodetection:** Zero-configuration automatic detection of the external WAN interface for NAT on the server side, and the physical default gateway on the client side.
* **High Performance:** Achieves bare-metal speeds and minimal overhead by leveraging native **TUN/TAP** interfaces on Linux and the robust **Wintun** driver on Windows.

---

## 🛠️ Architecture & CLI Flags

### 🐧 Server Mode (Linux)

When executed on a Linux environment, the binary acts as a full-featured VPN gateway, automatically configuring IP forwarding and NAT rules.

```bash
sudo ./liteobfsvpn [flags]

```

| Flag | Description | Default Value |
| --- | --- | --- |
| `-listen` | IP address and port to bind for incoming client connections | `0.0.0.0:443` |
| `-egress` | External network interface used for NAT routing | *Autodetected* |
| `-key` | Unique shared secret/cryptographic key for authorization | *Built-in default key* |

*Example usage:*

```bash
sudo ./liteobfsvpn -listen "0.0.0.0:8443" -key "my_super_secret_crypto_key_2026"

```

---

### 🪟 Client Mode (Windows)

When executed on Windows, the binary provisions a virtual network adapter and securely routes all global system traffic through the obfuscated tunnel.

> ⚠️ **Important:** To run the client successfully, the **`wintun.dll`** file must reside in the exact same directory as your executable (`liteobfsvpn.exe`).

```cmd
liteobfsvpn.exe [flags]

```

| Flag | Description | Default Value |
| --- | --- | --- |
| `-server` | Remote IP address and port of your VPN server | *Required parameter* |
| `-gw` | Physical gateway IP address (router local IP) | *Autodetected* |
| `-dns` | Custom DNS server to provision for the tunnel interface | `1.1.1.1` (Cloudflare) |
| `-key` | Unique shared cryptographic key matching the server | *Built-in default key* |

*Example usage (Run as Administrator):*

```cmd
liteobfsvpn.exe -server "203.0.113.50:8443" -key "my_super_secret_crypto_key_2026" -dns "8.8.8.8"

```

---

## 🚀 Quick Start

### Step 1: Deploy Server (Linux)

1. Transfer or build the binary on your remote Linux VPS.
2. Fire up the listener:

```bash
   sudo ./liteobfsvpn -listen "0.0.0.0:443" -key "your_personal_key"

```

### Step 2: Configure Client (Windows)

1. Move the compiled `liteobfsvpn.exe` to your local machine.
2. Download the official `wintun.dll` matching your system architecture (x64/x86) and drop it next to `liteobfsvpn.exe`.
3. Open a Command Prompt (CMD) or PowerShell session **as Administrator** and execute:

```cmd
   liteobfsvpn.exe -server "YOUR_SERVER_IP:443" -key "your_personal_key"

```

4. Verify your secure routing via any public IP checking service.

---

## 🔒 Security & Traffic Stealthing

1. **High-Entropy Noise:** The obfuscation layer strips away standard protocol indicators. To any middleboxes or ISPs, the connection looks like an unclassifiable stream of random bytes, eliminating DPI fingerprinting signatures (unlike OpenVPN or WireGuard).
2. **Key Security:** If the `-key` flag is omitted, a hardcoded fallback token is utilized. For robust security, **always** enforce a custom, high-entropy key string.

---

## 🛠️ Compilation (Development)

The project natively supports seamless cross-compilation out of the box:

```bash
# Compile for Linux Server
GOOS=linux GOARCH=amd64 go build -o liteobfsvpn main.go

# Compile for Windows Client
GOOS=windows GOARCH=amd64 go build -o liteobfsvpn.exe main.go

```

---

## 📄 License & Disclaimer

### "AS IS" Disclaimer

This software is provided **"AS IS"**, without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no event shall the authors or copyright holders be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the software or the use or other dealings in the software.

### Unlicense / Public Domain

This is free and unencumbered software released into the public domain. Anyone is free to copy, modify, publish, use, compile, sell, or distribute this software, either in source code form or as a compiled binary, for any purpose, commercial or non-commercial, and by any means. For more information, please refer to the [Unlicense](https://unlicense.org/) details.

```

```
