"""out-lang CLI wrapper.

The `out` command: locates the compiled Go binary and runs it.
Installed via `pip install out-lang`, or locally via `pip install -e .`.
"""
import os
import sys
import subprocess
import shutil


def _find_out_binary():
    """Locate the out binary relative to this module, build it, or find on PATH."""
    pkg_dir = os.path.dirname(os.path.abspath(__file__))
    binary_name = "out.exe" if sys.platform == "win32" else "out"

    candidates = [
        os.path.join(pkg_dir, binary_name),
        os.path.join(pkg_dir, "bin", binary_name),
        os.path.join(os.path.dirname(pkg_dir), binary_name),
        os.path.join(os.path.dirname(pkg_dir), "out.exe"),
    ]
    for c in candidates:
        if os.path.isfile(c):
            return c

    # try PATH
    found = shutil.which("out") or shutil.which("out.exe")
    if found:
        return found

    # try to build from source (requires go)
    return _build_and_locate()


def _build_and_locate():
    """Compile the Go source into a local binary."""
    pkg_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = os.path.dirname(pkg_dir)

    # If we are the project root itself (editable install)
    if os.path.isdir(os.path.join(project_root, "cmd", "out")):
        root = project_root
    elif os.path.isdir(os.path.join(pkg_dir, "cmd", "out")):
        root = pkg_dir
    else:
        return None

    binary_name = "out.exe" if sys.platform == "win32" else "out"
    dest = os.path.join(pkg_dir, binary_name)

    try:
        subprocess.run(
            ["go", "build", "-o", dest, "./cmd/out"],
            cwd=root,
            check=True,
            capture_output=True,
        )
        if os.path.isfile(dest):
            return dest
    except Exception:
        pass
    return None


def main():
    """Entry point for the `out` CLI."""
    binary = _find_out_binary()
    if not binary:
        print(
            "error: out binary not found.\n"
            "Run from project root: go build -o out.exe ./cmd/out\n"
            "Or ensure `out` is on your PATH.",
            file=sys.stderr,
        )
        sys.exit(1)
    result = subprocess.run([binary] + sys.argv[1:])
    sys.exit(result.returncode)


if __name__ == "__main__":
    main()
