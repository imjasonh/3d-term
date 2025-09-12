# 3d-term

A 3D ASCII cube renderer written in Go that displays a spinning cube in the terminal.

## Features

- **3D Math**: Complete implementation of 3D vector and matrix operations
- **3D Rendering**: Perspective projection with proper depth buffering (Z-buffer)
- **ASCII Shading**: Uses an ASCII ramp (" .-:=+*#%@") for realistic lighting effects
- **Animation**: Smooth rotation on all three axes (X, Y, Z)
- **Backface Culling**: Only renders visible faces for better performance

## How to Run

1. Make sure you have Go installed (1.21 or later)
2. Clone this repository
3. Build and run the program:

```bash
go build -o cube main.go
./cube
```

Or run directly:

```bash
go run main.go
```

## How it Works

The renderer implements:

1. **Mathematical Foundation**: 
   - `Vector` struct for 3D points
   - `Matrix` struct for 4x4 transformations
   - Matrix multiplication and rotation functions for X, Y, Z axes

2. **Scene Definition**:
   - Hardcoded cube with 8 vertices and 12 triangular faces
   - FrameBuffer for storing the rendered image
   - ZBuffer for depth testing

3. **Rendering Pipeline**:
   - Apply rotation transformations to cube vertices
   - Project 3D coordinates to 2D screen space
   - Calculate surface normals for lighting
   - Render triangles with proper depth testing
   - Use ASCII characters based on lighting intensity

4. **Animation Loop**:
   - Clear the terminal screen
   - Update rotation angles
   - Render the current frame
   - Repeat at ~20 FPS

Press Ctrl+C to exit the animation.