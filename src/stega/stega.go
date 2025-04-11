package stega

import (
    "bytes"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "fmt"
    "golang.org/x/crypto/pbkdf2"
    "image"
    "image/color"
    "image/png"
    "os"
    "strings"
)

type StegaMachine struct {
    StartChar string
    StopChar  string
}

func NewStegaMachine() *StegaMachine {
    return &StegaMachine{
        StartChar: string(rune(2)),
        StopChar:  string(rune(3)),
    }
}

func (s *StegaMachine) EncryptAES(message, password string) (string, error) {
    salt := make([]byte, 16)
    iv := make([]byte, aes.BlockSize)
    rand.Read(salt)
    rand.Read(iv)

    key := pbkdf2.Key([]byte(password), salt, 100_000, 32, sha256.New)
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    plaintext := pad([]byte(message))
    ciphertext := make([]byte, len(plaintext))
    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(ciphertext, plaintext)

    return "AES" + hex.EncodeToString(salt) + hex.EncodeToString(iv) + hex.EncodeToString(ciphertext), nil
}

func (s *StegaMachine) DecryptAES(ciphertextHex, password string) (string, error) {
    salt, _ := hex.DecodeString(ciphertextHex[3:35])
    iv, _ := hex.DecodeString(ciphertextHex[35:67])
    ct, _ := hex.DecodeString(ciphertextHex[67:])

    key := pbkdf2.Key([]byte(password), salt, 100_000, 32, sha256.New)
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    pt := make([]byte, len(ct))
    mode := cipher.NewCBCDecrypter(block, iv)
    mode.CryptBlocks(pt, ct)

    pt, err = unpad(pt)
    if err != nil {
        return "", err
    }
    return string(pt), nil
}

func pad(buf []byte) []byte {
    padding := aes.BlockSize - len(buf)%aes.BlockSize
    padtext := bytes.Repeat([]byte{byte(padding)}, padding)
    return append(buf, padtext...)
}

func unpad(buf []byte) ([]byte, error) {
    length := len(buf)
    if length == 0 {
        return nil, errors.New("invalid padding")
    }
    padding := buf[length-1]
    if int(padding) > aes.BlockSize || int(padding) > length {
        return nil, errors.New("invalid padding")
    }
    return buf[:length-int(padding)], nil
}

func toBinary(s string) string {
    var b bytes.Buffer
    for _, c := range s {
        b.WriteString(fmt.Sprintf("%08b", c))
    }
    return b.String()
}

func fromBinary(bin string) string {
    var out bytes.Buffer
    for i := 0; i+8 <= len(bin); i += 8 {
        b := bin[i : i+8]
        out.WriteByte(byte(parseBin(b)))
    }
    return out.String()
}

func parseBin(b string) uint8 {
    var result uint8
    for _, c := range b {
        result = result << 1
        if c == '1' {
            result |= 1
        }
    }
    return result
}

func (s *StegaMachine) HybridEmbedMessage(inputPath, message, password string, useAES bool) (string, error) {
    f, err := os.Open(inputPath)
    if err != nil {
        return "", err
    }
    defer f.Close()

    img, _, err := image.Decode(f)
    if err != nil {
        return "", err
    }

    if useAES {
        message, err = s.EncryptAES(message, password)
        if err != nil {
            return "", err
        }
    }

    msg := s.StartChar + message + s.StopChar
    binMsg := toBinary(msg)

    bounds := img.Bounds()
    width, height := bounds.Max.X, bounds.Max.Y

    outImg := image.NewNRGBA(bounds)
    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            outImg.Set(x, y, img.At(x, y))
        }
    }

    bitIdx := 0
    for y := 0; y < height && bitIdx < len(binMsg); y++ {
        for x := 0; x < width && bitIdx < len(binMsg); x++ {
            r, g, b, a := outImg.At(x, y).RGBA()
            // Set red channel
            cr := setLSB(uint8(r>>8), binMsg[bitIdx])
            bitIdx++

            // Check if we still have bits for green channel
            cg := uint8(g >> 8)
            if bitIdx < len(binMsg) {
                cg = setLSB(uint8(g>>8), binMsg[bitIdx])
                bitIdx++
            }

            // Check if we still have bits for blue channel
            cb := uint8(b >> 8)
            if bitIdx < len(binMsg) {
                cb = setLSB(uint8(b>>8), binMsg[bitIdx])
                bitIdx++
            }

            outImg.Set(x, y, color.NRGBA{cr, cg, cb, uint8(a >> 8)})
        }
    }

    if bitIdx < len(binMsg) {
        return "", errors.New("image not large enough to hold message")
    }

    outPath := "embedded_output.png"
    outFile, err := os.Create(outPath)
    if err != nil {
        return "", err
    }
    defer outFile.Close()

    err = png.Encode(outFile, outImg)
    return outPath, err
}

func setLSB(value byte, bit byte) byte {
    if bit == '1' {
        return value | 1
    }
    return value &^ 1
}

func getLSB(value byte) byte {
    return value & 1
}

func (s *StegaMachine) HybridExtractMessage(inputPath, password string) (string, error) {
    f, err := os.Open(inputPath)
    if err != nil {
        return "", err
    }
    defer f.Close()

    img, _, err := image.Decode(f)
    if err != nil {
        return "", err
    }

    bounds := img.Bounds()
    width, height := bounds.Max.X, bounds.Max.Y

    var bits strings.Builder
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            r, g, b, _ := img.At(x, y).RGBA()
            bits.WriteByte('0' + getLSB(uint8(r>>8)))
            bits.WriteByte('0' + getLSB(uint8(g>>8)))
            bits.WriteByte('0' + getLSB(uint8(b>>8)))
        }
    }

    msg := fromBinary(bits.String())
    start := strings.Index(msg, s.StartChar)
    stop := strings.Index(msg, s.StopChar)

    if start == -1 || stop == -1 || stop <= start {
        return "", errors.New("no valid hidden message found")
    }

    content := msg[start+1 : stop]
    if strings.HasPrefix(content, "AES") {
        if password == "" {
            return "PASSWORD_REQUIRED", nil
        }
        decrypted, err := s.DecryptAES(content, password)
        if err != nil {
            return "DECRYPTION_FAILED", nil
        }
        return decrypted, nil
    }

    return content, nil
}
