---
name: database-erd
description: Generates a PlantUML ERD diagram from the database schema and renders it to a PNG. Runs `just test::database-erd` to produce docs/erd.puml, then converts it to a PNG image. Use when asked to generate or visualize the database entity-relationship diagram.
---

# Database ERD Generator

Generates a PlantUML ERD diagram from the database schema by running the project's `just` recipe, then renders the resulting `.puml` file to a PNG.

## Setup

Ensure the required tools are available:

```bash
java -version 2>&1 | head -1
test -f ~/plantuml.jar && echo "plantuml.jar found" || echo "MISSING: ~/plantuml.jar"
which just && echo "just found" || echo "MISSING: just"
which feh && echo "feh found (preview)" || echo "feh not found (optional)"
```

## Workflow

### 1. Generate the ERD from the Database

Run the `just` recipe that connects to the database and runs `planter` to produce the PlantUML ERD:

```bash
cd <project-root> && just test::database-erd
```

This produces `Dev/docs/erd.puml` containing the PlantUML entity-relationship diagram of the database schema.

### 2. Render the Diagram to PNG

Use the render script to convert the `.puml` file to a PNG:

```bash
SKILL_DIR="$(dirname "$(readlink -f "$0")")"
"$SKILL_DIR/scripts/render.sh" /home/rodrigo/Documents/pvapi/Dev/docs/erd.puml
```

The script runs `java -jar ~/plantuml.jar -v -tpng` and opens the resulting PNG with `feh` for preview.

### Troubleshooting

- **`plantuml.jar` not found**: Place it at `~/plantuml.jar` or set `PLANTUML_JAR`
- **`java` missing**: Install a JDK/JRE
- **`just` not found**: Install with `cargo install just` or your package manager
- **Render errors**: Check the verbose output for PlantUML syntax issues
- **Database unreachable**: Ensure the test database container can start and is reachable
