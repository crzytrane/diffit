import { useCallback } from 'react';

export function UploadButton({ formRef, setDiffImageSrc }: { formRef: React.RefObject<HTMLFormElement>; setDiffImageSrc: React.Dispatch<React.SetStateAction<string>>; }) {
    const handleUpload = useCallback(async () => {
        if (formRef.current === null) {
            console.error("formRef.current is null");
            return;
        }

        const formData = new FormData(formRef.current);
        console.log("formData", formData.get("file-base"), formData.get("file-other"));

        fetch('http://localhost:4007/api/files', { method: 'POST', body: formData })
            .then(response => response.arrayBuffer())
            .then(buffer => {
                const base64String = btoa(
                    new Uint8Array(buffer)
                        .reduce((data, byte) => data + String.fromCharCode(byte), "")
                );
                setDiffImageSrc(`data:image/png;base64,${base64String}`);
            }
            ).catch(error => {
                console.error("error", error);
            });
    }, []);

    return (
        <button type="button" onClick={handleUpload}>Upload</button>
    );
}
