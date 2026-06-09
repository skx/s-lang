;;; s-lang.el --- mode for editing s-lang files

;; Copyright (C) 2026 Steve Kemp

;; Author: Steve Kemp <steve@steve.fi>
;; Keywords: languages
;; Version: 0.0.1

;;; Commentary:

;; Provides support for editing s-lang scripts with full support for
;; font-locking, but no special keybindings, or indentation handling.

;;;; Enabling:

;; Add the following to your .emacs file

;; (require 's-lang)
;; (setq auto-mode-alist (append '(("\\.in$" . s-lang-mode)) auto-mode-alist)))



;;; Code:

(defvar s-lang-keywords
  '(
    "else"
    "for"
    "function"
    "if"
    "let"
    "return"
    "while"
    ))

;; The functions from the standard-library.
(defvar s-lang-stdlib
  '(
    "argc"
    "argv"
    "exit"
    "false"
    "getc"
    "getenv"
    "malloc"
    "print"
    "putc"
    "rand"
    "readfile"
    "sleep"
    "sqrt"
    "str2float"
    "str2int"
    "strcat"
    "strcmp"
    "strdup"
    "strlen"
    "true"
    "type"
    ))


(defvar s-lang-font-lock-defaults
  `((
     ("\"\\.\\*\\?" . font-lock-string-face)
     (";\\|,\\|=" . font-lock-keyword-face)
     ( ,(regexp-opt s-lang-keywords 'words) . font-lock-builtin-face)
     ( ,(regexp-opt s-lang-stdlib 'words) . font-lock-function-name-face)
     )))

(define-derived-mode s-lang-mode prog-mode "s-lang sources"
  "s-lang-mode is a major mode for editing s-lang scripts"

  ;, comments
  (setq-local comment-start "#")
  (setq-local comment-start-skip "#+[\t ]*")

  ;; syntax
  (setq font-lock-defaults s-lang-font-lock-defaults)

  ;; comments
  (modify-syntax-entry ?# "<" s-lang-mode-syntax-table)
  (modify-syntax-entry ?\n ">" s-lang-mode-syntax-table)

  ;; blocks
  (modify-syntax-entry ?\{ "(}" s-lang-mode-syntax-table)
  (modify-syntax-entry ?\} "){" s-lang-mode-syntax-table)
  (modify-syntax-entry ?\( "()" s-lang-mode-syntax-table)
  )

(provide 's-lang)
