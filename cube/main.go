package main

import (
	"fmt"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/exp/gl/glutil"
)

type Shape struct {
	buf     gl.Buffer
	texture gl.Texture
}

type Shader struct {
	program      gl.Program
	vertCoord    gl.Attrib
	vertTexCoord gl.Attrib
	projection   gl.Uniform
	view         gl.Uniform
	model        gl.Uniform
}

type Engine struct {
	shader   Shader
	shape    Shape
	touchLoc geom.Point
	started  time.Time
	images   *glutil.Images
	fps      *debug.FPS
}

func (e *Engine) Start(glctx gl.Context) {
	var err error

	e.shader.program, err = LoadProgram(glctx, "shader.v.glsl", "shader.f.glsl")
	if err != nil {
		panic(fmt.Sprintln("LoadProgram failed:", err))
	}

	e.shape.buf = glctx.CreateBuffer()
	glctx.BindBuffer(gl.ARRAY_BUFFER, e.shape.buf)
	glctx.BufferData(gl.ARRAY_BUFFER, EncodeObject(cubeData), gl.STATIC_DRAW)

	e.shader.vertCoord = glctx.GetAttribLocation(e.shader.program, "vertCoord")
	e.shader.vertTexCoord = glctx.GetAttribLocation(e.shader.program, "vertTexCoord")

	e.shader.projection = glctx.GetUniformLocation(e.shader.program, "projection")
	e.shader.view = glctx.GetUniformLocation(e.shader.program, "view")
	e.shader.model = glctx.GetUniformLocation(e.shader.program, "model")

	e.shape.texture, err = LoadTexture(glctx, "gopher.png")
	if err != nil {
		panic(fmt.Sprintln("LoadTexture failed:", err))
	}

	e.started = time.Now()

	e.images = glutil.NewImages(glctx)
	e.fps = debug.NewFPS(e.images)
}

func (e *Engine) Stop(glctx gl.Context) {
	glctx.DeleteProgram(e.shader.program)
	glctx.DeleteBuffer(e.shape.buf)
	e.images.Release()
}


func (e *Engine) Draw(glctx gl.Context, c size.Event) {
	since := time.Now().Sub(e.started)

	glctx.Enable(gl.DEPTH_TEST)
	glctx.DepthFunc(gl.LESS)

	glctx.ClearColor(0, 0, 0, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
	glctx.Clear(gl.DEPTH_BUFFER_BIT)

	glctx.UseProgram(e.shader.program)

	m := mgl.Perspective(0.785, float32(c.WidthPt/c.HeightPt), 0.1, 10.0)
	glctx.UniformMatrix4fv(e.shader.projection, m[:])

	eye := mgl.Vec3{3, 3, 3}
	center := mgl.Vec3{0, 0, 0}
	up := mgl.Vec3{0, 1, 0}

	m = mgl.LookAtV(eye, center, up)
	glctx.UniformMatrix4fv(e.shader.view, m[:])

	m = mgl.HomogRotate3D(float32(since.Seconds()), mgl.Vec3{0, 1, 0})
	glctx.UniformMatrix4fv(e.shader.model, m[:])

	glctx.BindBuffer(gl.ARRAY_BUFFER, e.shape.buf)

	coordsPerVertex := 3
	texCoordsPerVertex := 2
	vertexCount := len(cubeData) / (coordsPerVertex + texCoordsPerVertex)

	glctx.EnableVertexAttribArray(e.shader.vertCoord)
	glctx.VertexAttribPointer(e.shader.vertCoord, coordsPerVertex, gl.FLOAT, false, 20, 0) // 4 bytes in float, 5 values per vertex

	glctx.EnableVertexAttribArray(e.shader.vertTexCoord)
	glctx.VertexAttribPointer(e.shader.vertTexCoord, texCoordsPerVertex, gl.FLOAT, false, 20, 12)

	glctx.BindTexture(gl.TEXTURE_2D, e.shape.texture)

	glctx.DrawArrays(gl.TRIANGLES, 0, vertexCount)

	glctx.DisableVertexAttribArray(e.shader.vertCoord)

	e.fps.Draw(c)
}

var cubeData = []float32{
	//  X, Y, Z, U, V
	// Bottom
	-1.0, -1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,

	// Top
	-1.0, 1.0, -1.0, 0.0, 0.0,
	-1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, -1.0, 1.0, 0.0,
	-1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, 1.0, 1.0, 1.0,

	// Front
	-1.0, -1.0, 1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 0.0, 0.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,
	1.0, -1.0, 1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,

	// Back
	-1.0, -1.0, -1.0, 0.0, 0.0,
	-1.0, 1.0, -1.0, 0.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	-1.0, 1.0, -1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 1.0, 1.0,

	// Left
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0,
	-1.0, -1.0, -1.0, 0.0, 0.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0,

	// Right
	1.0, -1.0, 1.0, 1.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	1.0, 1.0, -1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 0.0, 1.0,
}

func main() {
	e := Engine{}
	app.Main(func(a app.App) {
		var glctx gl.Context
		var c size.Event
		for eve := range a.Events() {
			switch eve := a.Filter(eve).(type) {
				case lifecycle.Event:
				switch eve.Crosses(lifecycle.StageVisible) {
					case lifecycle.CrossOn:
						glctx, _ = eve.DrawContext.(gl.Context)
					e.Start(glctx)
					case lifecycle.CrossOff:
					e.Stop(glctx)
						glctx = nil
				}
				case size.Event:
				c = eve
				e.touchLoc = geom.Point{c.WidthPt / 2, c.WidthPt / 2}
				case paint.Event:
					if glctx == nil || eve.External {
						continue
					}
				e.Draw(glctx, c)
				a.Publish()
				a.Send(paint.Event{})
				case touch.Event:
				// e.touchLoc = geom.Point{eve.X / 2, eve.Y / 2}
			}
		}
	})
}
