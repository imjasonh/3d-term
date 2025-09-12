package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Vector represents a 3D vector/point
type Vector struct {
	X, Y, Z float64
}

// Matrix represents a 4x4 matrix for 3D transformations
type Matrix struct {
	M [4][4]float64
}

// FrameBuffer represents the screen buffer
type FrameBuffer struct {
	Width, Height int
	Buffer        [][]rune
}

// ZBuffer represents the depth buffer
type ZBuffer struct {
	Width, Height int
	Buffer        [][]float64
}

// Face represents a triangular face of the cube
type Face struct {
	V1, V2, V3 int // indices into vertex array
}

// Cube vertices (8 corners of a unit cube centered at origin)
var cubeVertices = []Vector{
	{-1, -1, -1}, // 0
	{1, -1, -1},  // 1
	{1, 1, -1},   // 2
	{-1, 1, -1},  // 3
	{-1, -1, 1},  // 4
	{1, -1, 1},   // 5
	{1, 1, 1},    // 6
	{-1, 1, 1},   // 7
}

// Cube faces (12 triangles making 6 faces)
var cubeFaces = []Face{
	// Front face
	{0, 1, 2}, {0, 2, 3},
	// Back face
	{4, 6, 5}, {4, 7, 6},
	// Left face
	{0, 3, 7}, {0, 7, 4},
	// Right face
	{1, 5, 6}, {1, 6, 2},
	// Top face
	{3, 2, 6}, {3, 6, 7},
	// Bottom face
	{0, 4, 5}, {0, 5, 1},
}

// ASCII characters for shading (darkest to brightest)
var asciiRamp = " .-:=+*#%@"

// NewFrameBuffer creates a new frame buffer
func NewFrameBuffer(width, height int) *FrameBuffer {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
	}
	return &FrameBuffer{Width: width, Height: height, Buffer: buffer}
}

// NewZBuffer creates a new z-buffer
func NewZBuffer(width, height int) *ZBuffer {
	buffer := make([][]float64, height)
	for i := range buffer {
		buffer[i] = make([]float64, width)
	}
	return &ZBuffer{Width: width, Height: height, Buffer: buffer}
}

// Clear clears the frame buffer with spaces
func (fb *FrameBuffer) Clear() {
	for y := 0; y < fb.Height; y++ {
		for x := 0; x < fb.Width; x++ {
			fb.Buffer[y][x] = ' '
		}
	}
}

// Clear clears the z-buffer with far plane values
func (zb *ZBuffer) Clear() {
	for y := 0; y < zb.Height; y++ {
		for x := 0; x < zb.Width; x++ {
			zb.Buffer[y][x] = math.Inf(1) // Far plane
		}
	}
}

// SetPixel sets a pixel in the frame buffer if it passes the depth test
func (fb *FrameBuffer) SetPixel(x, y int, char rune, z float64, zb *ZBuffer) {
	if x >= 0 && x < fb.Width && y >= 0 && y < fb.Height {
		if z < zb.Buffer[y][x] {
			fb.Buffer[y][x] = char
			zb.Buffer[y][x] = z
		}
	}
}

// Print prints the frame buffer to stdout
func (fb *FrameBuffer) Print() {
	for y := 0; y < fb.Height; y++ {
		for x := 0; x < fb.Width; x++ {
			fmt.Print(string(fb.Buffer[y][x]))
		}
		fmt.Println()
	}
}

// Identity returns a 4x4 identity matrix
func Identity() Matrix {
	return Matrix{M: [4][4]float64{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}}
}

// RotationX returns a rotation matrix around the X axis
func RotationX(angle float64) Matrix {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Matrix{M: [4][4]float64{
		{1, 0, 0, 0},
		{0, cos, -sin, 0},
		{0, sin, cos, 0},
		{0, 0, 0, 1},
	}}
}

// RotationY returns a rotation matrix around the Y axis
func RotationY(angle float64) Matrix {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Matrix{M: [4][4]float64{
		{cos, 0, sin, 0},
		{0, 1, 0, 0},
		{-sin, 0, cos, 0},
		{0, 0, 0, 1},
	}}
}

// RotationZ returns a rotation matrix around the Z axis
func RotationZ(angle float64) Matrix {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Matrix{M: [4][4]float64{
		{cos, -sin, 0, 0},
		{sin, cos, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}}
}

// Multiply multiplies two 4x4 matrices
func (m1 Matrix) Multiply(m2 Matrix) Matrix {
	var result Matrix
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				result.M[i][j] += m1.M[i][k] * m2.M[k][j]
			}
		}
	}
	return result
}

// TransformVector applies a matrix transformation to a vector
func (m Matrix) TransformVector(v Vector) Vector {
	x := m.M[0][0]*v.X + m.M[0][1]*v.Y + m.M[0][2]*v.Z + m.M[0][3]
	y := m.M[1][0]*v.X + m.M[1][1]*v.Y + m.M[1][2]*v.Z + m.M[1][3]
	z := m.M[2][0]*v.X + m.M[2][1]*v.Y + m.M[2][2]*v.Z + m.M[2][3]
	w := m.M[3][0]*v.X + m.M[3][1]*v.Y + m.M[3][2]*v.Z + m.M[3][3]
	
	// Perspective division
	if w != 0 {
		return Vector{x / w, y / w, z / w}
	}
	return Vector{x, y, z}
}

// Project projects a 3D point to 2D screen coordinates
func Project(v Vector, width, height int, distance float64) (int, int, float64) {
	// Simple perspective projection
	if v.Z <= 0 {
		v.Z = 0.1 // Prevent division by zero
	}
	
	x := (v.X * distance / v.Z) + float64(width)/2
	y := (-v.Y * distance / v.Z) + float64(height)/2 // Flip Y for screen coordinates
	
	return int(x), int(y), v.Z
}

// CalculateNormal calculates the normal vector of a triangle face
func CalculateNormal(v1, v2, v3 Vector) Vector {
	// Calculate two edges of the triangle
	edge1 := Vector{v2.X - v1.X, v2.Y - v1.Y, v2.Z - v1.Z}
	edge2 := Vector{v3.X - v1.X, v3.Y - v1.Y, v3.Z - v1.Z}
	
	// Cross product to get normal
	normal := Vector{
		edge1.Y*edge2.Z - edge1.Z*edge2.Y,
		edge1.Z*edge2.X - edge1.X*edge2.Z,
		edge1.X*edge2.Y - edge1.Y*edge2.X,
	}
	
	// Normalize
	length := math.Sqrt(normal.X*normal.X + normal.Y*normal.Y + normal.Z*normal.Z)
	if length > 0 {
		normal.X /= length
		normal.Y /= length
		normal.Z /= length
	}
	
	return normal
}

// DrawTriangle draws a filled triangle using scanline algorithm
func DrawTriangle(fb *FrameBuffer, zb *ZBuffer, v1, v2, v3 Vector, x1, y1, z1, x2, y2, z2, x3, y3, z3 int, char rune) {
	// Simple approach: for each pixel in bounding box, check if inside triangle
	minX := min(x1, min(x2, x3))
	maxX := max(x1, max(x2, x3))
	minY := min(y1, min(y2, y3))
	maxY := max(y1, max(y2, y3))
	
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if x >= 0 && x < fb.Width && y >= 0 && y < fb.Height {
				// Barycentric coordinate test
				if isInsideTriangle(x, y, x1, y1, x2, y2, x3, y3) {
					// Interpolate Z value
					z := interpolateZ(x, y, x1, y1, float64(z1), x2, y2, float64(z2), x3, y3, float64(z3))
					fb.SetPixel(x, y, char, z, zb)
				}
			}
		}
	}
}

// isInsideTriangle checks if a point is inside a triangle using barycentric coordinates
func isInsideTriangle(px, py, x1, y1, x2, y2, x3, y3 int) bool {
	denominator := (y2-y3)*(x1-x3) + (x3-x2)*(y1-y3)
	if denominator == 0 {
		return false
	}
	
	a := float64((y2-y3)*(px-x3)+(x3-x2)*(py-y3)) / float64(denominator)
	b := float64((y3-y1)*(px-x3)+(x1-x3)*(py-y3)) / float64(denominator)
	c := 1 - a - b
	
	return a >= 0 && b >= 0 && c >= 0
}

// interpolateZ interpolates the Z value for a point inside a triangle
func interpolateZ(px, py, x1, y1 int, z1 float64, x2, y2 int, z2 float64, x3, y3 int, z3 float64) float64 {
	denominator := (y2-y3)*(x1-x3) + (x3-x2)*(y1-y3)
	if denominator == 0 {
		return z1
	}
	
	a := float64((y2-y3)*(px-x3)+(x3-x2)*(py-y3)) / float64(denominator)
	b := float64((y3-y1)*(px-x3)+(x1-x3)*(py-y3)) / float64(denominator)
	c := 1 - a - b
	
	return a*z1 + b*z2 + c*z3
}

// clearScreen clears the terminal screen
func clearScreen() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	width := 80
	height := 24
	distance := 100.0
	
	fb := NewFrameBuffer(width, height)
	zb := NewZBuffer(width, height)
	
	fmt.Println("3D ASCII Cube Renderer")
	fmt.Println("Press Ctrl+C to exit...")
	time.Sleep(2 * time.Second)
	
	angleX := 0.0
	angleY := 0.0
	angleZ := 0.0
	
	for {
		// Clear buffers
		fb.Clear()
		zb.Clear()
		
		// Create rotation matrix
		rotX := RotationX(angleX)
		rotY := RotationY(angleY)
		rotZ := RotationZ(angleZ)
		
		// Combine rotations
		transform := rotX.Multiply(rotY).Multiply(rotZ)
		
		// Transform vertices
		transformedVertices := make([]Vector, len(cubeVertices))
		for i, vertex := range cubeVertices {
			// Scale up the cube and move it away from camera
			scaled := Vector{vertex.X * 2, vertex.Y * 2, vertex.Z * 2 + 8}
			transformedVertices[i] = transform.TransformVector(scaled)
		}
		
		// Render faces
		for _, face := range cubeFaces {
			v1 := transformedVertices[face.V1]
			v2 := transformedVertices[face.V2]
			v3 := transformedVertices[face.V3]
			
			// Calculate normal for backface culling and shading
			normal := CalculateNormal(v1, v2, v3)
			
			// Simple backface culling (if normal points away from camera)
			if normal.Z < 0 {
				continue
			}
			
			// Project to 2D
			x1, y1, z1 := Project(v1, width, height, distance)
			x2, y2, z2 := Project(v2, width, height, distance)
			x3, y3, z3 := Project(v3, width, height, distance)
			
			// Simple lighting calculation (dot product with light direction)
			lightDir := Vector{0, 0, -1} // Light coming from camera
			intensity := math.Abs(normal.X*lightDir.X + normal.Y*lightDir.Y + normal.Z*lightDir.Z)
			
			// Map intensity to ASCII character
			charIndex := int(intensity * float64(len(asciiRamp)-1))
			if charIndex >= len(asciiRamp) {
				charIndex = len(asciiRamp) - 1
			}
			char := rune(asciiRamp[charIndex])
			
			// Draw the triangle
			DrawTriangle(fb, zb, v1, v2, v3, x1, y1, int(z1), x2, y2, int(z2), x3, y3, int(z3), char)
		}
		
		// Clear screen and display
		clearScreen()
		fb.Print()
		
		// Update rotation angles
		angleX += 0.05
		angleY += 0.03
		angleZ += 0.02
		
		// Small delay for animation
		time.Sleep(50 * time.Millisecond)
	}
}