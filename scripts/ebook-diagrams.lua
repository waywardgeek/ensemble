-- Some ebook renderers crop Graphviz SVGs at an incorrect scale. Select
-- generated raster companions only for these diagrams; print retains vectors.
local diagrams = {
  ["figures/common-ownership.svg"] = true,
  ["figures/common-context.svg"] = true,
  ["figures/common-contents.svg"] = true,
}

function Image(image)
  if diagrams[image.src] then
    image.src = image.src:gsub("%.svg$", ".png")
    image.attributes.width = "100%"
    image.attributes.height = nil
  end
  return image
end
