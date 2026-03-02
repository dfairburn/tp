" ftdetect/tp.vim - Filetype detection for tp template files
" Detects tp templates by checking if the file resides within the templates
" directory configured in ~/.tp/config.yml (templatesDirectoryPath).

function! s:GetTpTemplatesDir() abort
  " Try to read the tp config file and extract templatesDirectoryPath
  let l:config_paths = [
        \ expand('~/.tp/config.yml'),
        \ expand('~/.tp/config.yaml'),
        \ expand('~/.tp/config/config.yml'),
        \ expand('~/.tp/config/config.yaml'),
        \ ]

  for l:cfg in l:config_paths
    if filereadable(l:cfg)
      for l:line in readfile(l:cfg)
        let l:match = matchstr(l:line, 'templatesDirectoryPath:\s*"\?\zs[^"]\+\ze"\?')
        if !empty(l:match)
          " Expand ~ to home directory
          return substitute(l:match, '^\~', expand('~'), '')
        endif
      endfor
    endif
  endfor

  " Fall back to the default templates directory
  return expand('~/.tp/templates')
endfunction

function! s:DetectTpTemplate() abort
  let l:templates_dir = s:GetTpTemplatesDir()
  " Normalise both paths (resolve symlinks, trailing slashes)
  let l:templates_dir = fnamemodify(l:templates_dir, ':p')
  let l:file_dir = fnamemodify(expand('%:p'), ':p')

  if stridx(l:file_dir, l:templates_dir) == 0
    set filetype=tp
  endif
endfunction

autocmd BufRead,BufNewFile *.yml,*.yaml call s:DetectTpTemplate()
