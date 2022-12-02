#pragma once
#ifndef TKP_GB_APU_H
#define TKP_GB_APU_H
#include <SDL2/SDL.h>
#include <queue>
#include <GameboyTKP/gb_apu_ch.h>
namespace TKPEmu::Gameboy::Devices {
    // This class is solely for sound output and is not needed to pass sound
    // emulation tests.
    // All computation for this class happens in gb_bus and gb_apu_ch
    class APU {
    public:
        APU(ChannelArrayPtr channel_array_ptr, uint8_t& NR52);
        ~APU();
        void InitSound();
        void Update(int clk);
        inline void QueueSamples() {
            #ifndef NOAUDIO
            SDL_QueueAudio(device_id_, &samples_[0], sizeof(samples_));
            #endif
        }
        inline bool IsQueueEmpty() {
            #ifndef NOAUDIO
            return SDL_GetQueuedAudioSize(device_id_) < 100;
            #else
            return false;
            #endif
        }
        bool UseSound = false;
    private:
        #ifndef NOAUDIO
        SDL_AudioDeviceID device_id_;
        #endif
        std::array<int16_t, 512> samples_;
        size_t sample_index_ = 0;
        uint8_t& NR52_;
        ChannelArrayPtr channel_array_ptr_;
        bool init_ = false;
    };
}
#endif