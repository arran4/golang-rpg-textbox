# rpgtextbox AI Agent Skill

This skill teaches AI agents how to correctly interact with the `rpgtextbox` CLI tool.

## Tool Overview
`rpgtextbox` is a command-line application designed to generate RPG-style text boxes. It supports static image generation as well as animated text rendering.

## Key Commands and Usage
The main command you will use is `rpgtextbox generate`.

### Generating a Text Box
To generate an image with text, you must at least specify the input text (via file or standard input). Other important flags:
* `--width` (default 600) and `--height` (default 150): The dimensions of the generated image.
* `--themedir` (default "./theme"): The directory containing the theme assets. Note: If the default directory does not exist in the current working directory, generation will fail unless you specify a valid path.
* `--text` (default ""): Text file to use, or `-` for standard input.
* `--out` (default "out-"): Prefix for the generated output files (e.g., `out-01.png`).

Example usage:
```bash
echo "Hello, world!" | rpgtextbox generate --text - --out mybox
```

### Animations
If you supply the `--animation` flag (e.g. `--animation letter-by-letter-animation`), `rpgtextbox` will output an animated GIF instead of a PNG, and will append `-animated` to the output file prefix.

Example:
```bash
echo "Hello, world!" | rpgtextbox generate --text - --animation letter-by-letter-animation --out animated-box
```

### Chevrons and Avatars
* Use `--chevron <type>` to add a continue indicator (e.g., `end-of-text-chevron`).
* Use `--avatar-pos <type>` and `--avatar-scale <type>` to include an avatar. Note: This requires an `avatar.png` in the theme directory.

## Common Traps
* **Theme Directory**: Ensure you are running the tool from a directory that contains a `theme/` folder or explicitly pass `--themedir` to a valid location. The default `theme/fromdirpng` requires `frame.png`, `chevron.png`, and `avatar.png`.
* **Animations Output**: Animated files use the prefix and add `-animated.gif`, whereas static pages add `-XX.png` (where XX is the page number).
