/**
 * This is unused, but I will keep it because it might be useful.
 * I was going to add video streaming but my code wouldn't work
 */

const getDuration = (file: File) => {
  const url = URL.createObjectURL(file);
  return new Promise<number>((resolve) => {
    const audio = document.createElement("audio");
    audio.muted = true;
    const source = document.createElement("source");
    source.src = url; //--> blob URL
    audio.preload = "metadata";
    audio.appendChild(source);
    audio.onloadedmetadata = function () {
      resolve(audio.duration);
    };
  });
};

export default getDuration;
