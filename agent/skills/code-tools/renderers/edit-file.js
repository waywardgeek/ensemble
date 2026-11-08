import { fileDiff } from "../../shared/file-diff.js";
export default function render(document, preview, directory, file) {
    return fileDiff(document, file);
}
