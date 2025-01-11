type DiffImageCompareViewerProps = {
    diffImageSrc: string;
    baseImageSrc: string;
    otherImageSrc: string;
}

export function DiffImageCompareViewer({ diffImageSrc, baseImageSrc, otherImageSrc }: DiffImageCompareViewerProps) {
    return <div>
        <img src={diffImageSrc} alt="Diff"></img>
        <img src={baseImageSrc} alt="Base image to compare" />
        <img src={otherImageSrc} alt="Other image to compare" />
    </div>
}