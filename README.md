# Steganography Tool

This is a GUI steganography tool that supports hybrid embedding techniques, combining **adaptive LSB** and **DCT-based** methods to hide secret messages within images. Optional **AES encryption** is also supported to secure the hidden message. The tool preserves image quality and **EXIF metadata**, allowing the output image to retain key info from the original.

**Supported formats:** `.jpg`, `.jpeg`, `.png`, `.bmp`, `.webp`

---

## Motivation

This project started in April 2025, inspired by a LinkedIn discussion with **Santeri Kallio**, where the topic was:

> *"ChatGPT:llä tehtyihin kuviin lisätään dataa mikä paljastaa että tekoäly on generoinut sen."*  
> *(Posted on 4.4.2025)*

While **C2PA** provides robust AI image attribution, I wanted to explore the broader potential of hiding extra data inside images—whether it’s secret messages, watermarks, or provenance data. This tool demonstrates how hidden content can be embedded *without visibly degrading* the image.

---

## Features

### 🧠 Hybrid Embedding Techniques

- **Adaptive LSB:** Dynamically adjusts bit depth based on local image complexity.
- **DCT-based embedding:** Uses mid-frequency DCT coefficients for more robust hiding.

### 🔐 Optional AES Encryption

- Encrypt messages with AES (CBC mode)
- Uses `PBKDF2` for key derivation
- Random salt and IV are automatically managed

### 🧾 Metadata Preservation

- Original **EXIF metadata** is preserved in the final output image

### 🖥️ GUI Interface

- Built using **Fyne**
- Simple interface for file selection, encryption/decryption, and message entry

---

## Requirements (for source version)

Go version 1.20+

If package manager has 1.20+ version, it can be installed:
`sudo apt install go -y`

Installing from tar package:
```
wget https://golang.org/dl/go1.24.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
export PATH=/usr/local/go1.24.2/go/bin:$PATH
echo "alias go='/usr/local/go1.23.0/go/bin/go'" >> ~/.bashrc
source ~/.bashrc
```

---

## Installation (from source)

```
git clone https://github.com/divergentti/steganography.git
cd steganography/src

go mod init stegago
go mod tidy
go build -o stegago.go
```

The code is modular and split into:
- GUI layer
- Encryption/decryption engine

---

## Usage

Run the main application:

`./stegago.go`

### Encryption Mode

1. Select image or folder
2. Enter message
3. (Optional) Enable AES and set password
4. Click **Encrypt**

### Decryption Mode

1. Select encrypted image
2. (If used) Enter password
3. Click **Decrypt**

✅ The tool uses checksums for message integrity and preserves EXIF metadata during processing.

---

## License

This project is licensed under the **MIT License**.
