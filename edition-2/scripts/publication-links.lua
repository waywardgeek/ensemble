-- Local repository files are not bundled into PDF/EPUB. Keep chapter anchors
-- and external URLs intact, but send source/evidence links to the public repo.
local function publication_link(link, directory)
  local target = link.target
  if target:match('^#') or target:match('^[%a][%w+.-]*:') or target:match('^//') then
    return link
  end
  local path, suffix = target:match('^([^?#]*)(.*)$')
  local parts = {}
  if not target:match('^/') then
    for part in directory:gmatch('[^/]+') do table.insert(parts, part) end
  end
  for part in path:gmatch('[^/]+') do
    if part == '..' then
      table.remove(parts)
    elseif part ~= '.' then
      table.insert(parts, part)
    end
  end
  local kind = path:match('/$') and 'tree' or 'blob'
  link.target = 'https://github.com/waywardgeek/ensemble/' .. kind .. '/main/' ..
    table.concat(parts, '/'):gsub(' ', '%%20') .. suffix
  return link
end

-- Keep the published book as the default; isolated editions supply their own
-- repository directory. Read metadata before walking links (Meta runs too late
-- for a top-level Link callback).
function Pandoc(doc)
  local directory = doc.meta['publication-directory']
  directory = directory and pandoc.utils.stringify(directory) or 'book'
  return doc:walk({Link = function(link) return publication_link(link, directory) end})
end
