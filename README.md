# ndsemu

[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/rasky/ndsemu?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Status

This emulator is still in **DEVELOPMENT PHASE**.

Some games work and are playable, others have broken graphics, others crash
almost immediately.

### What is emulated

 * ARM
   * Most opcodes implemented in both ARM and Thumb mode
   * Disassemly for debugging support
   * Correct cycle counting
   * Correct handling of miasligned memory addresses
   * Preliminar JIT (not fully working yet)
 * 2D: BG layers
   * Text mode (16/256 colors, scrolling)
   * Affine modes (16bit bgmap, 8bit bitmap, direct bitmap)
   * Different map sizes
   * Priorities
   * Extended palettes
 * 2D: OBJ (sprites)
   * Normal sprites
   * Different sizes
   * Extended palettes
   * Bitmaps
   * Affine (rotozoom)
 * 2D: advanced modes
   * VRAM display mode
 * 2D: misc features
   * Capturing: only basic support (normal BG+OBJ capture)
   * Master brightness
   * Window (OBJ and BG)
   * Color special effects
 * 3D: geometry processor
   * Most commands implemented
   * Accurate timing
 * 3D: rasterizer
   * Quadrangle splitting
   * Backface culling
   * Triangle rasterization
   * All different texture formats
   * Texture perspective correction
   * Clipping
   * Lighting and materials (with bugs...)
   * Toon shading
 * Sound
   * PCM channels
   * Noise

### What is NOT emulated

 * 2D
   * Mosaic
 * 3D
   * Tons of small fixes
   * Light perspective corrections
   * Edge marking
   * Fog
 * Sound
   * Capture (also for reverbs) 
   * Mic input
 * Emulator features
   * Savestates
   * Replays
 
## How to compile

To compile, you must clone into a `ndsemu` subdirectory:

    git clone https://github.com/rasky/ndsemu $GOPATH/src/ndsemu
    cd $GOPATH/src/ndsemu
    go get
    go build

## BIOS

You need access to an official NDS BIOS and firmware. Put them within a "bios" subdirectory, like this:

    ndsemu
      |---bios
           |---- firmware.bin
           |---- biosnds7.rom
           |---- biodnds9.rom

## Run it

At this point, you can just run it with:

    ./ndsemu <path-to-your-rom-file>


