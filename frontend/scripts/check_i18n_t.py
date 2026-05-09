from pathlib import Path
root = Path(__file__).resolve().parents[1] / "pages"
for p in sorted(root.rglob("*.vue")):
    t = p.read_text(encoding="utf-8")
    if "t('errors.api" in t and "useI18n" not in t:
        print(p)
