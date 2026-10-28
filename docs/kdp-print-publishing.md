# KDP print publishing: 7 × 10 paperback and hardcover

## Interior: shared layout, separate eligibility checks

Paperback and hardcover can use the same **7 × 10 inch, no-bleed interior PDF**
when its page count, margins and other settings meet both products' requirements.
Hardcover case-wrap and hinge allowances belong to the **cover**, not the
interior. Do not add cover-wrap margins to `book/print-header.html`.

Current print CSS: mirrored 0.875-inch inner and 0.5-inch outer margins, with
0.75 inches at top and bottom. KDP's shared margin table specifies a minimum
0.75-inch gutter for 301–500 pages and 0.875 inches for 501–700 pages; minimum
outer margins for a no-bleed interior are 0.25 inches. Recheck the applicable
row whenever pagination changes.

At this cover-preparation check, `book/The Art of Building AI Coding Agents -
Print Interior.pdf` contains **496 physical pages**, all **504 × 720 points**.
This is within KDP's listed **75–550-page** hardcover range for 7 × 10 inches.
The older 577-page publication records describe earlier artifacts, not the
current PDF. These checks alone do not constitute complete KDP acceptance.

Upload the **Print Interior.pdf**, not the cover-prefixed reading PDF, as the
manuscript. No interior rebuild or script change was needed for this cover task.

## Hardcover front image for Cover Creator

Repository asset:

- [`book/ensemble-edition-2-hardcover-front-7x10.png`](../book/ensemble-edition-2-hardcover-front-7x10.png)

Local delivery directory: `/Users/bill/Downloads/AI Coding Agents - Hardcover Front/`.

- `hardcover-front-7x10-400dpi.jpg`: convenient Cover Creator upload.
- `hardcover-front-7x10-400dpi.png`: lossless alternative, identical to the repository asset.
- `preparation.json`: source/output SHA-256 hashes and processing parameters.
- `README.txt`: short upload notes.

Both images are **2800 × 4000 pixels, RGB, 400 DPI**, without transparency.
The source is Bill's 4096 × 6144 image:
`/Users/bill/Downloads/The Art of Building AI Coding Agents (1)_upscayl_4x_digital-art-4x.png`.
It was not modified.

The entire source is proportionally downsampled to 2400 × 3600 pixels and placed
at (200, 200) on the new canvas. A 200-pixel surround extends the nearest edge
colors, blurred with radius 28; a 28-pixel feather joins the artwork edge to the
surround. There is no crop, stretching, mirrored lettering or regenerated text.
The source is assumed to be sRGB. No CMYK conversion is applied to these
Cover Creator image assets; print color must be checked in a proof.

OCR of the delivered JPEG detected all six original text lines. Approximate
minimum text clearances on the 7 × 10 canvas are 0.96 inches left, 0.93 right,
1.13 top and 0.87 bottom. OCR is a mechanical placement check, not a visual
quality review or a measurement of final Cover Creator placement. The OCR
engine reads the artwork's `AI` as `Al`; the image lettering was not changed.

### Upload procedure and limits

1. In the **hardcover** setup, select 7 × 10 inches and the final manuscript,
   paper/ink options and page count.
2. Launch **Cover Creator** and use the JPEG or PNG as the **front image** in a
   suitable layout. KDP lists JPG/JPEG, PNG and TIF/TIFF as accepted image types.
3. The artwork already includes title, subtitle and author. Avoid duplicate
   template text over it; inspect the resulting front, spine and back.
4. Inspect the actual hardcover safe lines. Scaling or cropping by a template
   can consume the extra space. Keep lettering and important artwork out of
   the hinge and wrapped areas; adjust placement rather than stretching.
5. Preview the complete cover and interior, resolve warnings, and order a proof
   before publication. No live Cover Creator placement or physical proof has
   been performed for this asset.

**The 7:10 image is a front-design asset, not a calculated hardcover board or
case-wrap file.** Cover Creator determines the full layout. Do not use it through
the complete-cover upload option or stretch it across the spine and back.
Do not treat the earlier `AI Coding Agents - 7x10 Cover` paperback bleed files as
hardcover-ready assets.

For a manually assembled cover, download the **hardcover** template from KDP's
calculator using the actual trim, page count and paper/ink choices. Build a
single back/spine/front PDF against that template, including wrap and hinge
allowances. KDP specifies a 0.4-inch hinge area and hardcover-specific safe
margins. Its hardcover page currently prints a wrap dimension as `0.51” (15 mm)`;
those units are not equivalent. Do not derive a production template from that
ambiguous conversion—use the calculator's dimensioned template.

The often-cited **1600 × 2560** cover recommendation is for a Kindle eBook,
not a 7 × 10 print front image. KDP's print-cover minimum is **300 DPI** at the
final placed size. An upscale's pixel count alone does not establish sharpness.

## Official references

- [Hardcover cover requirements](https://kdp.amazon.com/help/topic/GDTKFJPNQCBTMRV6)
- [Cover Creator and supported images](https://kdp.amazon.com/help/topic/G201113520)
- [Cover calculator/templates](https://kdp.amazon.com/cover-calculator)
- [Trim, bleed and interior margin table](https://kdp.amazon.com/help/topic/GVBQ3CMEQW3W2VL6)
- [Hardcover manuscript formatting](https://kdp.amazon.com/help/topic/GKYZRXFBZH2LDXAK)

No KDP-specific skill was found in the active repository or the personal skill
locations checked during this task; this guide is the maintained publishing note.
No upload, publication or modification of the manuscript/build scripts was
performed as part of this cover preparation.
