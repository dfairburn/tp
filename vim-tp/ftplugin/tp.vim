" ftplugin/tp.vim - Filetype settings and JSON validation for tp templates

if exists('b:did_ftplugin')
  finish
endif
let b:did_ftplugin = 1

" Sensible defaults for editing tp templates
setlocal commentstring=#\ %s
setlocal comments=:#
setlocal expandtab
setlocal shiftwidth=2
setlocal softtabstop=2

" ---------------------------------------------------------------------------
" :TpValidate - Validate JSON in the body section of a tp template
"
" Extracts the body block, replaces {{ ... }} template expressions with
" valid JSON placeholder strings, then validates the result using either
" jq or python3 -m json.tool (whichever is available).
" Errors are reported in the quickfix list with correct line numbers.
" ---------------------------------------------------------------------------

function! s:FindValidator() abort
  if executable('jq')
    return 'jq .'
  elseif executable('python3')
    return 'python3 -m json.tool'
  elseif executable('python')
    return 'python -m json.tool'
  else
    return ''
  endif
endfunction

function! s:TpValidateBody() abort
  let l:validator = s:FindValidator()
  if empty(l:validator)
    echoerr 'TpValidate: No JSON validator found. Install jq or python3.'
    return
  endif

  let l:lines = getline(1, '$')
  let l:body_start = -1
  let l:body_indent = ''
  let l:body_lines = []

  " Find the 'body:' line
  let l:idx = 0
  for l:line in l:lines
    let l:idx += 1
    if l:line =~# '^\s*body:'
      let l:body_start = l:idx
      break
    endif
  endfor

  if l:body_start < 0
    echo 'TpValidate: No body section found.'
    return
  endif

  " Collect indented lines after body:
  let l:first_body_line = -1
  let l:idx = l:body_start + 1
  while l:idx <= len(l:lines)
    let l:line = l:lines[l:idx - 1]

    " Skip empty lines at the start
    if l:first_body_line < 0
      if l:line =~# '^\s*$'
        let l:idx += 1
        continue
      endif
      " First non-empty line determines the indent level
      let l:body_indent = matchstr(l:line, '^\s\+')
      if empty(l:body_indent)
        " Not indented = no body content or next top-level key
        break
      endif
      let l:first_body_line = l:idx
    endif

    " If line is not indented (and not empty), we've left the body block
    if l:line !~# '^\s' && l:line !~# '^\s*$'
      break
    endif

    " Strip the body indentation prefix
    let l:stripped = substitute(l:line, '^\s\{' . len(l:body_indent) . '}', '', '')
    call add(l:body_lines, l:stripped)
    let l:idx += 1
  endwhile

  if empty(l:body_lines)
    echo 'TpValidate: Body section is empty.'
    return
  endif

  " Replace {{ ... }} template expressions with valid JSON placeholder strings.
  " This prevents template syntax from causing false JSON validation errors.
  " Two passes:
  "   1. Replace "{{ ... }}" (already quoted) with "__TP_TEMPLATE__" (keep quotes)
  "   2. Replace remaining bare {{ ... }} with "__TP_TEMPLATE__" (add quotes)
  let l:processed = []
  for l:bline in l:body_lines
    " Pass 1: template inside quotes -> keep the quotes, replace content
    let l:clean = substitute(l:bline, '"\zs{{\s*.\{-}\s*}}\ze"', '__TP_TEMPLATE__', 'g')
    " Pass 2: bare template (not already inside quotes) -> wrap in quotes
    let l:clean = substitute(l:clean, '{{\s*.\{-}\s*}}', '"__TP_TEMPLATE__"', 'g')
    call add(l:processed, l:clean)
  endfor

  let l:json_text = join(l:processed, "\n")

  " Write to a temp file and validate
  let l:tmpfile = tempname()
  call writefile(split(l:json_text, "\n"), l:tmpfile)

  let l:output = system(l:validator . ' < ' . shellescape(l:tmpfile))
  let l:exit_code = v:shell_error

  call delete(l:tmpfile)

  if l:exit_code == 0
    echo 'TpValidate: Body JSON is valid.'
    cclose
    return
  endif

  " Parse errors and map line numbers back to the original file
  let l:qflist = []
  for l:errline in split(l:output, "\n")
    " Try to extract a line number from the error message
    " jq format: "parse error (Invalid numeric literal at line X, column Y)"
    " python format: "... line X column Y ..."
    let l:lnum_match = matchstr(l:errline, '\cline \zs\d\+')
    let l:lnum = 0
    if !empty(l:lnum_match)
      " Map back to original file line number
      let l:lnum = str2nr(l:lnum_match) + l:first_body_line - 1
    endif

    call add(l:qflist, {
          \ 'filename': expand('%'),
          \ 'lnum': l:lnum,
          \ 'text': l:errline,
          \ 'type': 'E',
          \ })
  endfor

  if !empty(l:qflist)
    call setqflist(l:qflist)
    copen
    echohl ErrorMsg
    echo 'TpValidate: Body JSON has errors. See quickfix list.'
    echohl None
  else
    echohl ErrorMsg
    echo 'TpValidate: ' . l:output
    echohl None
  endif
endfunction

command! -buffer TpValidate call s:TpValidateBody()

" Optional: validate on write
" Uncomment the following line to auto-validate the body JSON on save:
" autocmd BufWritePost <buffer> TpValidate

let b:undo_ftplugin = 'delcommand TpValidate'
          \ . '| setlocal commentstring< comments< expandtab< shiftwidth< softtabstop<'
