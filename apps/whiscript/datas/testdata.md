
- -t 180：先頭から 180秒（=3分）だけ 処理 
- -ac 1：モノラルに変換 
- -c:a pcm_s16le：WAV 16-bit PCM で出力

```bash
  INPUT=/Users/hnao/audiohijack/20251002/20251002\ 1008\ Recording\ 3.mp3
  OUTPUT=hama.wav
  ffmpeg -i $INPUT -t 180 -ac 1 -c:a pcm_s16le $OUTPUT
```

```bash
  source ~/venvs/whisper_env312/bin/activate
```


```bash
  IN=hama.wav
whisperx $IN --model large-v3 --device cpu --compute_type int8 --threads 8 --language ja --interpolate_method linear --chunk_size 20 --vad_method silero --vad_onset 0.6 --vad_offset 0.35 --align_model "jonatasgrosman/wav2vec2-large-xlsr-53-japanese" --output_dir out_hama --output_format json
```