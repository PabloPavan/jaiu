(() => {
  function photoUpload(el) {
    const dataset = (el && el.dataset) || {};
    const initialObjectKey = dataset.photoObjectKey || "";
    const initialUrl = dataset.photoUrl || "";

    return {
      photoObjectKey: initialObjectKey,
      previewUrl: initialUrl,
      error: "",
      async handleFile(event) {
        const file = event && event.target && event.target.files && event.target.files[0];
        if (!file) {
          return;
        }
        this.error = "";

        if (this.previewUrl && this.previewUrl.startsWith("blob:")) {
          URL.revokeObjectURL(this.previewUrl);
        }
        this.previewUrl = URL.createObjectURL(file);
      },
      clear() {
        if (this.previewUrl && this.previewUrl.startsWith("blob:")) {
          URL.revokeObjectURL(this.previewUrl);
        }
        this.photoObjectKey = "";
        this.previewUrl = "";
        this.error = "";
        if (this.$refs && this.$refs.photoInput) {
          this.$refs.photoInput.value = "";
        }
      },
    };
  }

  window.photoUpload = photoUpload;
})();
