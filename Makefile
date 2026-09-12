.PHONY: paper paper-clean

PAPER_OUTDIR ?= paper

paper:
	./scripts/build-paper.sh --outdir "$(PAPER_OUTDIR)"

paper-clean:
	rm -f paper/geometry-of-work.pdf paper/*.aux paper/*.bbl paper/*.blg paper/*.log paper/*.out paper/*.toc paper/*.fls paper/*.fdb_latexmk
