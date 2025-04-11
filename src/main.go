package main

import (
    "fmt"
    "image"
    "stegago/stega"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/widget"
)

func main() {
    a := app.NewWithID("com.example.stegago")
    w := a.NewWindow("Steganography Tool v0.1.0")
    w.Resize(fyne.NewSize(900, 600))

    machine := stega.NewStegaMachine()

    messageEntry := widget.NewEntry()
    messageEntry.SetPlaceHolder("Enter your secret message")

    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Password (optional)")

    encryptCheck := widget.NewCheck("Enable AES Encryption", func(b bool) {
        passwordEntry.Disable()
        if b {
            passwordEntry.Enable()
        }
    })

    encryptMode := widget.NewRadioGroup([]string{"Encrypt", "Decrypt"}, nil)
    encryptMode.Horizontal = true
    encryptMode.Selected = "Encrypt"

    selectedPath := widget.NewEntry()
    selectedPath.Disable()

    imgPreview := canvas.NewText("Image Preview", nil)
    previewContainer := container.NewMax(imgPreview)

    selectFileBtn := widget.NewButton("Browse", func() {
        dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
            if reader != nil {
                selectedPath.SetText(reader.URI().Path())
                img, _, err := image.Decode(reader)
                if err == nil {
                    imgCanvas := canvas.NewImageFromImage(img)
                    imgCanvas.FillMode = canvas.ImageFillContain
                    previewContainer.Objects = []fyne.CanvasObject{imgCanvas}
                    previewContainer.Refresh()
                }
            }
        }, w)
    })

    actionBtn := widget.NewButton("Encrypt", func() {
        mode := encryptMode.Selected
        path := selectedPath.Text
        message := messageEntry.Text
        password := passwordEntry.Text
        useAES := encryptCheck.Checked

        if path == "" {
            dialog.ShowError(fmt.Errorf("Please select a file"), w)
            return
        }

        if mode == "Encrypt" && message == "" {
            dialog.ShowError(fmt.Errorf("Message is required for encryption"), w)
            return
        }

        if useAES && password == "" {
            dialog.ShowError(fmt.Errorf("Password is required if AES is enabled"), w)
            return
        }

        if mode == "Encrypt" {
            outPath, err := machine.HybridEmbedMessage(path, message, password, useAES)
            if err != nil {
                dialog.ShowError(err, w)
            } else {
                dialog.ShowInformation("Success", fmt.Sprintf("Saved to %s", outPath), w)
            }
        } else {
            msg, err := machine.HybridExtractMessage(path, password)
            if err != nil {
                dialog.ShowError(err, w)
            } else if msg == "PASSWORD_REQUIRED" {
                dialog.ShowInformation("Password Needed", "Decryption password required.", w)
            } else if msg == "DECRYPTION_FAILED" {
                dialog.ShowInformation("Failed", "Incorrect password or corrupted data.", w)
            } else {
                dialog.ShowInformation("Decrypted Message", msg, w)
            }
        }
    })

    encryptMode.OnChanged = func(s string) {
        actionBtn.SetText(s)
        if s == "Decrypt" {
            messageEntry.Hide()
        } else {
            messageEntry.Show()
        }
    }

    controls := container.NewVBox(
        encryptMode,
        container.NewBorder(nil, nil, nil, selectFileBtn, selectedPath),
        messageEntry,
        encryptCheck,
        passwordEntry,
        actionBtn,
    )

    mainLayout := container.NewHSplit(
        container.NewVBox(layout.NewSpacer(), controls, layout.NewSpacer()),
        previewContainer,
    )
    mainLayout.Offset = 0.4

    w.SetContent(mainLayout)
    w.ShowAndRun()
}
