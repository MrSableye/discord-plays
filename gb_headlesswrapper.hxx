#ifndef GB_HEADLESSWRAPPER_HXX
#define GB_HEADLESSWRAPPER_HXX
#include <GameboyTKP/gb_addresses.h>
#include <GameboyTKP/gb_cpu.h>
#include <GameboyTKP/gb_ppu.h>
#include <GameboyTKP/gb_bus.h>
#include <GameboyTKP/gb_timer.h>
#include <GameboyTKP/gb_apu.h>
#include <GameboyTKP/gb_apu_ch.h>

#define GifFrameCount 60

using Frame = std::array<uint8_t, 320 * 288 * 4>;
enum class Command {
    Reset,
    Left,
    Up,
    Right,
    Down,
    A,
    B,
    Start,
    Select,
    Frame,
    Screenshot,
    Second,
    Exit,
    Save,
    Load,
    ReadSingle,
    ReadString,
    Error,
    GetParty,
    GetBalls,
    GetTrainer,
    GetMap,
    StartSave,
    Spam,
    GetGif,
};

class Gameboy {
private:
    using GameboyPalettes = std::array<std::array<float, 3>,4>;
    using GameboyKeys = std::array<uint32_t, 4>;
    using CPU = TKPEmu::Gameboy::Devices::CPU;
    using PPU = TKPEmu::Gameboy::Devices::PPU;
    using APU = TKPEmu::Gameboy::Devices::APU;
    using ChannelArrayPtr = TKPEmu::Gameboy::Devices::ChannelArrayPtr;
    using ChannelArray = TKPEmu::Gameboy::Devices::ChannelArray;
    using Bus = TKPEmu::Gameboy::Devices::Bus;
    using Timer = TKPEmu::Gameboy::Devices::Timer;
    using Cartridge = TKPEmu::Gameboy::Devices::Cartridge;
public:
    Gameboy(std::string path);
    Gameboy(const Gameboy&) = default;
    void ExecuteCommand(Command command);
    void SetValue(std::string val) { value_ = val; }
    void SetMemory(uint16_t addr, uint8_t val) { bus_.Write(addr, val); }
    std::string GetRes() { return res_; }
private:
    ChannelArrayPtr channel_array_ptr_;
    Bus bus_;
    APU apu_;
    PPU ppu_;
    Timer timer_;
    CPU cpu_;
    GameboyKeys direction_keys_;
    GameboyKeys action_keys_;
    uint8_t& interrupt_flag_;
    std::array<std::unique_ptr<Frame>, GifFrameCount> last_minute_frames_;
    int last_minute_frame_index_ = 0;

    void reset();
    void frame();
    void update();
    void screenshot();
    void save();
    void load();
    void get_gif();
    void get_party();
    void get_balls();
    void get_trainer();
    void get_map();

    std::string poke_to_ascii(uint8_t* name, int size);
    std::string poke_type_to_name(uint8_t type);
    std::string poke_item_to_name(uint8_t type);

    int frame_clk_ = 0;
    std::string value_;
    std::string res_;
};
#endif