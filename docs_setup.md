# Documentation Development Setup

This project uses **MkDocs** to build and serve the documentation.

Follow these steps to create a Python virtual environment, install the required dependencies, and run the documentation locally.

---

## 1. Create a Virtual Environment

Create a virtual environment in the project directory:

```bash
python -m venv .venv
```

---

## 2. Activate the Virtual Environment (Fish shell)

If you are using the **Fish shell**, activate the environment with:

```bash
source .venv/bin/activate.fish
```

For reference:

| Shell      | Command                          |
| ---------- | -------------------------------- |
| bash / zsh | `source .venv/bin/activate`      |
| fish       | `source .venv/bin/activate.fish` |

---

## 3. Install Documentation Dependencies

Install the dependencies listed in `requirements.txt`:

```bash
pip install -r requirements.txt
```

---

## 4. Update Dependencies

If you add new MkDocs plugins or themes, update the requirements file:

```bash
pip freeze > requirements.txt
```

Then commit the updated file.

---

## 5. Run the Documentation Server

Start the local documentation server:

```bash
mkdocs serve
```

By default it will be available at:

```
http://127.0.0.1:8000
```

The server supports **live reload**, so changes to the documentation will automatically refresh the page.

---

## 6. Build Static Documentation

To generate the static site used for deployment:

```bash
mkdocs build
```

The generated files will be placed in the `site/` directory.

---

## Directory Overview

```
docs/        → Markdown documentation
mkdocs.yml   → MkDocs configuration
site/        → Generated static site (after build)
requirements.txt → Python dependencies
```
