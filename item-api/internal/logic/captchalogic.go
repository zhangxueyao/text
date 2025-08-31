package logic

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zhangxueyao/item/item-api/internal/svc"
	"github.com/zhangxueyao/item/item-api/internal/types"
)

type CaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaLogic {
	return &CaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaptchaLogic) Captcha() (*types.CaptchaResp, error) {
	id := uuid.New().String()
	code := randomDigits(4)
	l.svcCtx.CaptchaStore.Set(id, code, 5*time.Minute)
	img := drawCaptcha(code)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return &types.CaptchaResp{CaptchaId: id, ImageBase64: encoded}, nil
}

func randomDigits(n int) string {
	digits := make([]byte, n)
	for i := 0; i < n; i++ {
		digits[i] = byte('0' + rand.Intn(10))
	}
	return string(digits)
}

// digit patterns 5x7
var digitPatterns = map[rune][7]string{
	'0': {"01110", "10001", "10011", "10101", "11001", "10001", "01110"},
	'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
	'2': {"01110", "10001", "00001", "00010", "00100", "01000", "11111"},
	'3': {"11110", "00001", "00001", "01110", "00001", "00001", "11110"},
	'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
	'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
	'6': {"00110", "01000", "10000", "11110", "10001", "10001", "01110"},
	'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9': {"01110", "10001", "10001", "01111", "00001", "00010", "01100"},
}

func drawCaptcha(code string) image.Image {
	scale := 4
	spacing := scale
	width := len(code)*(5*scale) + (len(code)-1)*spacing
	height := 7 * scale
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// white background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}
	for i, ch := range code {
		drawDigit(img, ch, i*(5*scale+spacing), scale)
	}
	return img
}

func drawDigit(img *image.RGBA, ch rune, offsetX int, scale int) {
	pattern, ok := digitPatterns[ch]
	if !ok {
		return
	}
	for y, row := range pattern {
		for x, c := range row {
			if c == '1' {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						img.Set(offsetX+x*scale+dx, y*scale+dy, color.Black)
					}
				}
			}
		}
	}
}
