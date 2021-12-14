package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/gift"
	"github.com/oakmound/oak/v3"
	"github.com/oakmound/oak/v3/alg"
	"github.com/oakmound/oak/v3/alg/floatgeom"
	"github.com/oakmound/oak/v3/alg/intgeom"
	"github.com/oakmound/oak/v3/debugstream"
	"github.com/oakmound/oak/v3/event"
	"github.com/oakmound/oak/v3/key"
	"github.com/oakmound/oak/v3/render"
	"github.com/oakmound/oak/v3/render/mod"
	"github.com/oakmound/oak/v3/scene"
	"golang.org/x/image/colornames"
)

var colors = []color.Color{
	colornames.Yellow,
	colornames.Red,
	colornames.Blue,
	colornames.Green,
	colornames.Indigo,
	colornames.Teal,
	colornames.Purple,
	colornames.Darkgreen,
	colornames.Limegreen,
	colornames.Magenta,
	colornames.Turquoise,
	colornames.Brown,
	colornames.Darkorange,
	colornames.Darkgray,
	colornames.Darkkhaki,
}

var options = []string{
	"Option 1",
	"Option 2",
	"Option 3",
	"Option 4",
	"Option 5",
}

const radius = 300.0

func main() {
	oak.AddScene("spinner", scene.Scene{
		Start: func(ctx *scene.Context) {
			rand.Seed(time.Now().Unix())

			if len(colors) > len(options) {
				colors = colors[:len(options)]
			}

			bkg, _ := render.LoadSprite("gameshow.png")
			bkg.Modify(mod.Resize(ctx.Window.Width(), ctx.Window.Height(), gift.CubicResampling))
			ctx.Window.(*oak.Window).SetBackground(bkg)

			totalSlices := 26.0
			degrees := 360 / totalSlices
			centerX, centerY := float64(ctx.Window.Width())/2, float64(ctx.Window.Height())/2

			centerX += 40

			p1 := floatgeom.Point2{0, 0}
			upAngle := floatgeom.AnglePoint(degrees / 2).MulConst(radius)
			downAngle := floatgeom.AnglePoint(-degrees / 2).MulConst(radius)
			p2 := p1.Add(upAngle)
			p3 := p1.Add(downAngle)

			poly := floatgeom.NewPolygon2(p1, p2, p3)
			polys := make([]render.Renderable, int(totalSlices))

			teamNameBackground := render.NewColorBoxR(300, 420, color.RGBA{180, 180, 180, 200})
			teamNameBackground.SetPos(15, 15)
			render.Draw(teamNameBackground, 1)

			for i, name := range options {
				c := colors[i%len(colors)]
				fnt, _ := render.DefaultFont().RegenerateWith(func(fg render.FontGenerator) render.FontGenerator {
					fg.Color = image.NewUniform(c)
					fg.Size = 28
					return fg
				})
				txt := fnt.NewText(name, 20, (float64(i)*28)+32).ToSprite()
				render.Draw(txt, 10)
			}
			for i := 0; i < int(totalSlices); i++ {
				c := colors[int(float64(i)*(float64(len(colors))/totalSlices))%len(colors)]
				poly = rotatePolyAroundPoint(poly, p1, degrees)
				polyR := render.NewPolygon(poly)
				polyR.ShiftX(centerX)
				polyR.ShiftY(centerY)
				polyR.Fill(c)
				polys[i] = polyR
			}
			comp := render.NewCompositeR(polys...)

			wheel := comp.ToSprite()
			wheel.Modify(mod.TrimColor(color.RGBA{0, 0, 0, 0}))
			w, h := wheel.GetDims()
			fmt.Println(w, h)

			wheelRotated := map[string]render.Modifiable{}
			for i := 0; i < 360; i++ {
				idx := strconv.Itoa(i)
				wCopy := wheel.Copy().Modify(mod.Rotate(float32(i)))
				wCopy.Modify(mod.TrimColor(color.RGBA{0, 0, 0, 0}))
				wCopy.Modify(mod.Resize(600, 600, mod.CubicResampling))
				fillRGBA(wCopy.GetRGBA())
				wheelRotated[idx] = wCopy
			}
			wheelSwitch := render.NewSwitch("0", wheelRotated)
			wheelSwitch.SetPos(centerX-float64(w/2), centerY-float64(h/2))
			render.Draw(wheelSwitch, 2)

			rotation := 0.0
			currentDegrees := 0.0
			event.GlobalBind("EnterFrame", func(c event.CID, i interface{}) int {
				if rotation > 0.2 {
					wheelSwitch.Set(strconv.Itoa(int(currentDegrees)))
					currentDegrees = currentDegrees + rotation
					if currentDegrees > 360 {
						currentDegrees = currentDegrees - 360
					}
					if rotation < 1 {
						rotation *= .995
					} else {
						rotation *= .99
					}
				}
				return 0
			})
			event.GlobalBind(key.Down+key.Spacebar, func(c event.CID, i interface{}) int {
				rotation += float64(rand.Intn(20) + 10)
				return 0
			})
			arrow, _ := render.LoadSprite("arrow.png")
			arrow.Modify(mod.FlipX)
			arrow.Modify(mod.Scale(.25, .25))
			arrow.ShiftX(centerX + 270)
			arrow.ShiftY(centerY - 100)
			render.Draw(arrow, 10)
			debugstream.AddCommand(debugstream.Command{
				Name: "add",
				Operation: func(s []string) string {
					if len(s) == 0 {
						return ""
					}
					options = append(options, strings.Join(s, " "))
					go ctx.Window.GoToScene("spinner")
					return ""
				},
			})
			debugstream.AddCommand(debugstream.Command{
				Name: "remove",
				Operation: func(s []string) string {
					if len(s) == 0 {
						return ""
					}
					name := strings.Join(s, " ")
					for i, t := range options {
						if t == name {
							options = append(options[:i], options[i+1:]...)
							go ctx.Window.GoToScene("spinner")
							return ""
						}
					}
					return ""
				},
			})
			c := render.NewCircle(colornames.White, 30, 0)
			fillRGBAWith(c.GetRGBA(), 30.0, colornames.White)
			cw, ch := c.GetDims()
			c.SetPos(centerX-float64(cw)/2, centerY-float64(ch)/2)
			render.Draw(c, 11)
		},
	})
	oak.Init("spinner", func(c oak.Config) (oak.Config, error) {
		c.EnableDebugConsole = true
		c.Title = "Wheel of Selection!"
		c.Screen.Width = 1280
		c.Screen.Height = 960
		return c, nil
	})
}

func rotatePolyAroundPoint(ply floatgeom.Polygon2, center floatgeom.Point2, rotation float64) floatgeom.Polygon2 {
	rads := alg.DegToRad * rotation
	cos := math.Cos(rads)
	sin := math.Sin(rads)
	for i, p := range ply.Points {
		dx := p.X() - center.X()
		dy := p.Y() - center.Y()

		p[0] = center.X() + ((dx * cos) - (dy * sin))
		p[1] = center.Y() + ((dx * sin) + (dy * cos))
		ply.Points[i] = p
	}
	return floatgeom.NewPolygon2(ply.Points[0], ply.Points[1], ply.Points[2], ply.Points[3:]...)
}

func fillRGBA(r *image.RGBA) {
	w, h := r.Bounds().Max.X, r.Bounds().Max.Y
	cX := w / 2
	cY := h / 2
	cp := intgeom.Point2{cX, cY}

	for x := 0; x < w; x++ {
	Y:
		for y := 0; y < h; y++ {
			p := intgeom.Point2{x, y}
			if p.Distance(cp) <= radius {
				c := r.RGBAAt(x, y)
				if c.A < 200 {
					for x2 := x - 3; x2 < x+4; x2++ {
						for y2 := y - 3; y2 < y+4; y2++ {
							c2 := r.RGBAAt(x2, y2)
							if c2.A == 255 {
								r.SetRGBA(x, y, c2)
								continue Y
							}
						}
					}
				}
			}
		}
	}
}

func fillRGBAWith(r *image.RGBA, radius float64, c2 color.RGBA) {
	w, h := r.Bounds().Max.X, r.Bounds().Max.Y
	cX := w / 2
	cY := h / 2
	cp := intgeom.Point2{cX, cY}

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			p := intgeom.Point2{x, y}
			if p.Distance(cp) <= radius {
				c := r.RGBAAt(x, y)
				if c.A < 200 {
					r.SetRGBA(x, y, c2)
				}
			}
		}
	}
}
