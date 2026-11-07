import { fileExcerpt } from "../../shared/file-excerpt.js";
export default function render(document, preview, directory, file) {
    return fileExcerpt(document, file, ["read"]);
}
