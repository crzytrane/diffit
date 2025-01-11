import { useCallback, useRef } from 'react';
import styles from './ImageUpload.module.css';

export function ImageUpload({ $name, imgSrc, setImageSrc }: UploadableImageProps) {
    const name = $name.toLowerCase();
    const inputRef = useRef<HTMLInputElement>(null);

    const handleImageChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (file === undefined) {
            setImageSrc("");
            return;
        }

        let reader = new FileReader();

        reader.onload = () => {
            setImageSrc(reader.result as string);
        };

        reader.readAsDataURL(file);
    }, []);

    const handleImageClick = useCallback(() => {
        if (inputRef.current === null) {
            console.error("inputRef.current is null");
            return;
        }

        inputRef.current.click();
    }, []);

    const hideStyles = { display: 'none' };
    const hasImage = imgSrc !== "";
    const inputStyles = hasImage ? hideStyles : {};

    return (
        <>
            {imgSrc && <img src={imgSrc} className={styles.image} alt="Image to diff" onClick={handleImageClick} />}
            <input style={hideStyles} type="file" name={`file-${name}`} aria-label={`${name} image upload`} onChange={handleImageChange} ref={inputRef} />
            <button type="button" className={styles.selectButton} style={inputStyles} onMouseDown={handleImageClick}>Select { }</button>
        </>
    );
}
