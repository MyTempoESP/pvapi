#!/bin/bash
# Render PlantUML diagrams and preview with feh.
# Usage: ./render.sh <file.puml> [file2.puml ...]
set -eu

PLANTUML_JAR="${PLANTUML_JAR:-$HOME/plantuml.jar}"

if [[ ! -f "$PLANTUML_JAR" ]]; then
    echo "ERROR: plantuml.jar not found at $PLANTUML_JAR" >&2
    echo "Set PLANTUML_JAR or place it at ~/plantuml.jar" >&2
    exit 1
fi

if command -v java &>/dev/null; then
	java -jar "$PLANTUML_JAR" -v -tpng "$@"
else
	echo "ERROR: missing java runtime" >&2
	exit 1
fi

PNG_FILES=()
for FILE in "$@"; do
    if [[ -f "$FILE" ]]; then
        BASE="${FILE%.*}"
        PNG="${BASE}.png"
        if [[ -f "$PNG" ]]; then
            PNG_FILES+=("$PNG")
        fi
    fi
done

if (( ${#PNG_FILES[@]} > 0 )); then
    if command -v feh &>/dev/null; then
        feh --geometry 800x600 "${PNG_FILES[@]}" &
    else
        echo "Generated: ${PNG_FILES[*]}"
        echo "(install feh for automatic preview)"
    fi
else
    echo "No valid .puml files were processed." >&2
    exit 1
fi
