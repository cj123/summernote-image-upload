export class SummernoteImageUploader {
    private $element: JQuery;
    private readonly uploadURL: string;
    private readonly opts: Summernote.Options;
    private readonly code?: string;

    public constructor(uploadURL: string, $element: JQuery, opts: Summernote.Options, code?: string) {
        this.uploadURL = uploadURL;
        this.$element = $element;
        this.opts = opts;
        this.code = code;
    }

    public render(): void {
        this.opts.callbacks = {
            onImageUpload: (files: FileList) => {
                // @ts-ignore
                for (let file of files) {
                    this.uploadFile(file);
                }
            }
        }

        this.$element.summernote(this.opts);

        if (this.code) {
            this.$element.summernote("code", this.code);
        }
    }

    private uploadFile(file: File) {
        let data = new FormData();
        data.append("image", file);

        $.ajax({
            url: this.uploadURL,
            type: "POST",
            data: data,
            contentType: false,
            processData: false,
            success: (url: string) => {
                this.$element.summernote("editor.insertImage", url);
            },
        })
    }
}
