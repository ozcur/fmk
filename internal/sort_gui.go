package internal

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/pixiv/go-libjpeg/jpeg"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

const (
	sGUIWidth  = 1200
	sGUIHeight = 1200

	infoNameStr       = "**Name**: %s"
	infoDimensionsStr = "**Dimensions**: %dx%d"
	infoIndexStr      = "**Index**: %d/%d"
)

var (
	fyneHotkeys = map[fyne.KeyName]int{
		fyne.Key1: 1,
		fyne.Key2: 2,
		fyne.Key3: 3,
		fyne.Key4: 4,
		fyne.Key5: 5,
		fyne.Key6: 6,
		fyne.Key7: 7,
		fyne.Key8: 8,
		fyne.Key9: 9,
	}
)

type sGUI struct {
	a          fyne.App
	w          fyne.Window
	sortPaths  []string
	imagePaths []string
	buttons    []*widget.Button
	imgB       *sortBrowser
}

func (g *sGUI) start() error {
	g.w.ShowAndRun()
	return nil
}

type sortBrowser struct {
	imagePaths     []string
	imagePathsIdx  int
	fCanvas        *canvas.Image
	infoName       *widget.RichText
	infoDimensions *widget.RichText
	infoIndex      *widget.RichText
	curImg         image.Image
}

func newSortBrowser(imagePaths []string) *sortBrowser {
	f, err := os.Open(imagePaths[0])
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open image file.")
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode image file.")
	}

	fCanvas := canvas.NewImageFromImage(img)
	fCanvas.FillMode = canvas.ImageFillContain

	ret := &sortBrowser{
		imagePaths:     imagePaths,
		imagePathsIdx:  0,
		fCanvas:        fCanvas,
		infoName:       widget.NewRichTextFromMarkdown("nil"),
		infoDimensions: widget.NewRichTextFromMarkdown("nil"),
		infoIndex:      widget.NewRichTextFromMarkdown("nil"),
		curImg:         img,
	}

	ret.refreshInfo()

	return ret
}

func (ib *sortBrowser) next() {
	ib.imagePathsIdx++
	log.Debug().Int("imagePathsIdx", ib.imagePathsIdx).
		Msg("imagePathsIdx incremented.")
	ib.refreshImg()
	ib.refreshInfo()
}

func (ib *sortBrowser) refreshImg() {
	openStart := time.Now()
	f, err := os.Open(ib.imagePaths[ib.imagePathsIdx])
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open image file.")
	}
	openDuration := time.Since(openStart)

	decodeStart := time.Now()

	// Large jpeg files decode very slow using stdlib.  Use
	// a different library to decode them.
	img := image.Image(nil)
	ext := strings.ToLower(filepath.Ext(ib.imagePaths[ib.imagePathsIdx]))
	if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(f, &jpeg.DecoderOptions{})
		log.Debug().Msg("decoded with 3rd party lib")
	} else {
		img, _, err = image.Decode(f)
		log.Debug().Msg("decoded with stdlib")
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode image file.")
	}

	decodeDuration := time.Since(decodeStart)

	setImgStart := time.Now()
	ib.curImg = img
	ib.fCanvas.Image = img
	setImgDuration := time.Since(setImgStart)

	refreshStart := time.Now()
	ib.fCanvas.Refresh()
	refreshDuration := time.Since(refreshStart)

	log.Debug().
		Str("openDuration", openDuration.String()).
		Str("decodeDuration", decodeDuration.String()).
		Str("setImgDuration", setImgDuration.String()).
		Str("refreshDuration", refreshDuration.String()).
		Msg("Refreshed image durations")
}

func (ib *sortBrowser) refreshInfo() {
	nameF := fmt.Sprintf(infoNameStr, filepath.Base(ib.imagePaths[ib.imagePathsIdx]))
	ib.infoName.ParseMarkdown(nameF)

	dimensionsF := fmt.Sprintf(infoDimensionsStr,
		ib.curImg.Bounds().Dx(),
		ib.curImg.Bounds().Dy(),
	)
	ib.infoDimensions.ParseMarkdown(dimensionsF)

	indexF := fmt.Sprintf(infoIndexStr, ib.imagePathsIdx+1, len(ib.imagePaths))
	ib.infoIndex.ParseMarkdown(indexF)
}

func (ib *sortBrowser) isMore() bool {
	return ib.imagePathsIdx < len(ib.imagePaths)-1
}

func (ib *sortBrowser) sort(fn fileHandleFunc, sortPath string) error {
	newPath := filepath.Join(sortPath, filepath.Base(ib.imagePaths[ib.imagePathsIdx]))
	return fn(ib.imagePaths[ib.imagePathsIdx], newPath)
}

func newSGUI(c *cli.Context, sortPaths []string, imagePaths []string) *sGUI {
	a := app.New()
	w := a.NewWindow("fmk - Sort")

	// Setup hotkey bar
	hotkeyBar := container.NewHBox()
	buttons := [](*widget.Button){}

	// Setup image
	imgB := newSortBrowser(imagePaths)

	// Setup info bar
	infoBar := container.NewHBox()
	infoBar.Add(imgB.infoName)
	infoBar.Add(imgB.infoDimensions)
	infoBar.Add(layout.NewSpacer())
	infoBar.Add(imgB.infoIndex)

	// Add quit button
	quitB := widget.NewButton("[ESC] Quit", func() {})
	hotkeyBar.Add(quitB)

	// Setup sorting buttons
	var fn fileHandleFunc
	if c.Bool("copy") {
		fn = fileHandlerCopy
		log.Debug().Msg("Using copy mode.")
	} else {
		fn = fileHandlerMove
		log.Debug().Msg("Using move mode.")
	}

	for i, sPath := range sortPaths {
		p := sPath // closure

		buttonName := fmt.Sprintf("[%d] %s", i+1, filepath.Base(sPath))
		hB := widget.NewButton(buttonName, func() {
			log.Info().Str("path", p).Msg("Sorting to folder.")
			err := imgB.sort(fn, p)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to sort image.")
			}
		})

		hotkeyBar.Add(hB)
		buttons = append(buttons, hB)
	}

	// Setup window
	content := container.NewBorder(hotkeyBar, infoBar, nil, nil, imgB.fCanvas)
	w.SetContent(content)
	w.Resize(fyne.NewSize(sGUIWidth, sGUIHeight))

	// Setup hotkeys
	w.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		if ke.Name == fyne.KeyEscape {
			w.Close()
		}

		if keyNum, ok := fyneHotkeys[ke.Name]; ok {
			log.Debug().Str("key", string(ke.Name)).
				Int("keyNum", keyNum).
				Msg("Hotkey pressed.")
			if keyNum > len(buttons) {
				return
			}
			buttons[keyNum-1].Tapped(nil)
			if imgB.isMore() {
				imgB.next()
			} else {
				w.Close()
			}
		}
	})

	return &sGUI{
		a:          a,
		w:          w,
		sortPaths:  sortPaths,
		imagePaths: imagePaths,
		buttons:    buttons,
		imgB:       imgB,
	}
}
