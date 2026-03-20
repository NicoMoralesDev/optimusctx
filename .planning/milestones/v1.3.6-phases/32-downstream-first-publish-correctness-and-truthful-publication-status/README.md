# Phase 32

Status: not planned

Title: Downstream first-publish correctness and truthful publication status

Scope:
- fix Homebrew and Scoop publish behavior when the generated file is new in an otherwise empty downstream repo
- repair downstream publication truth so `published` only means a real repo update happened
- add regression coverage for first-file publication paths
