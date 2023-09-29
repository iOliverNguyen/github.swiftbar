package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/fogleman/gg"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

var (
	red    = rgba(255, 78, 73, 255)
	green  = rgba(110, 175, 40, 255)
	yellow = rgba(231, 180, 71, 255)
	gray   = rgba(114, 114, 114, 255)
	violet = rgba(175, 73, 255, 255)
)

func rgba(r, g, b, a byte) color.RGBA {
	return color.RGBA{r, g, b, a}
}

func calcColor(status string, merged bool) color.Color {
	switch status {
	case "pass":
		if merged {
			return violet
		}
		return green
	case "fail":
		return red
	case "pending":
		return yellow
	default:
		return gray
	}
}

func renderStackImage(stack *Stack) image.Image {
	N := len(stack.TopPRs)
	W, H := (4+2)*N-2, 16
	img := image.NewRGBA(image.Rect(0, 0, W, H))

	for x, pr := range stack.TopPRs {
		x0, x1 := float64((4+2)*x), float64((4+2)*x+4)
		checks := pr.MostChecks()

		pass, fail, pending, skipping := 0, 0, 0, 0
		for _, ch := range checks {
			switch ch.Status {
			case "pass":
				pass++
			case "fail":
				fail++
			case "pending":
				pending++
			case "skipping":
				skipping++
			}
		}
		total := pass + fail + pending
		if total == 0 {
			total, skipping = 1, 1
		}

		debugf("--- CHECKS ---")
		debugYaml(checks)
		debugf("PR#%v pass=%v fail=%v pending=%v total=%v", pr.Number, pass, fail, pending, total)

		var dc *gg.Context
		y0, y1 := float64(0), float64(pass*H)/float64(total)
		dc = gg.NewContextForRGBA(img)
		if pr.Merged {
			dc.SetColor(violet)
		} else {
			dc.SetColor(green)
		}
		dc.DrawRectangle(x0, y0, x1-x0, y1-y0)
		dc.Fill()
		debugf("GREEN x0=%v x1=%v y0=%v y1=%v", x0, x1, y0, y1)

		y2 := y1 + float64(pending*H)/float64(total)
		dc = gg.NewContextForRGBA(img)
		dc.SetColor(yellow)
		dc.DrawRectangle(x0, y1, x1-x0, y2-y1)
		dc.Fill()
		debugf("YELLO x0=%v x1=%v y1=%v y2=%v", x0, x1, y1, y2)

		y3 := y2 + float64(skipping*H)/float64(total)
		y3 = math.Round(y3)
		dc = gg.NewContextForRGBA(img)
		dc.SetColor(gray)
		dc.DrawRectangle(x0, y2, x1-x0, y3-y2)
		dc.Fill()
		debugf("GRAY  x0=%v x1=%v y2=%v y3=%v", x0, x1, y2, y3)

		y4 := y3 + float64(fail*H)/float64(total)
		y4 = math.Round(y4)
		dc = gg.NewContextForRGBA(img)
		dc.SetColor(red)
		dc.DrawRectangle(x0, y3, x1-x0, y4-y3)
		dc.Fill()
		debugf("RED   x0=%v x1=%v y3=%v y4=%v", x0, x1, y3, y4)
	}
	debugImage("stack.png", img)
	return img
}

func renderPRImage(pr *PR) image.Image {
	checks := pr.ImportantChecks()
	N := len(checks)
	W, H := (4+2)*N-2, 16
	img := image.NewRGBA(image.Rect(0, 0, W, H))

	for x, ch := range checks {
		x0, x1 := float64((4+2)*x), float64((4+2)*x+4)
		y0, y1 := float64(0), float64(H)
		dc := gg.NewContextForRGBA(img)
		if ch.Status == "" {
			dc.SetColor(gray)
		} else {
			dc.SetColor(calcColor(ch.Status, pr.Merged))
		}
		dc.DrawRectangle(x0, y0, x1-x0, y1-y0)
		dc.Fill()
	}
	debugImage(fmt.Sprintf("pr%v.png", pr.Number), img)
	return img
}

func imageToBase64(img image.Image) []byte {
	b := &bytes.Buffer{}
	must(0, png.Encode(b, img))

	data, encoder := b.Bytes(), base64.StdEncoding
	out := make([]byte, encoder.EncodedLen(len(data)))
	base64.StdEncoding.Encode(out, data)
	return out
}

func debugImage(file string, img image.Image) {
	if verbosed {
		fullpath := pathLogDir + file
		b := &bytes.Buffer{}
		must(0, png.Encode(b, img))
		must(0, os.WriteFile(fullpath, b.Bytes(), 0644))
	}
}
