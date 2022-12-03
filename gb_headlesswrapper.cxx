#include "gb_headlesswrapper.hxx"
#define STB_IMAGE_WRITE_IMPLEMENTATION
#include "stb_image_write.h"
#include <fstream>
#include <vector>
#include <iostream>
#include <image_scale.hxx>
#define write32(data) {\
    uint32_t x = data; \
    ofs.write((char*)&x, sizeof(x));\
}

#define write16(data) {\
    uint16_t x = data; \
    ofs.write((char*)&x, sizeof(x));\
}

#define write8(data) {\
    uint8_t x = data; \
    ofs.write((char*)&x, sizeof(x));\
}

#define direction(index) bus_.DirectionKeys &= (~(1UL << index)); \
            interrupt_flag_ |= IFInterrupt::JOYPAD; \
            for (int i = 0; i < 5; i++) \
                ExecuteCommand(Command::Frame); \
            bus_.DirectionKeys |= (1UL << index); \
            ExecuteCommand(Command::Second);

#define action(index) bus_.ActionKeys &= (~(1UL << index)); \
            interrupt_flag_ |= IFInterrupt::JOYPAD; \
            for (int i = 0; i < 5; i++) \
                ExecuteCommand(Command::Frame); \
            bus_.ActionKeys |= (1UL << index); \
            ExecuteCommand(Command::Second);

Gameboy::Gameboy(std::string path) :
    channel_array_ptr_(std::make_shared<ChannelArray>()),
    bus_(channel_array_ptr_),
    apu_(channel_array_ptr_, bus_.GetReference(addr_NR52)),
    ppu_(bus_, &DrawMutex),
    timer_(channel_array_ptr_, bus_),
    cpu_(bus_, ppu_, apu_, timer_),
    interrupt_flag_(bus_.GetReference(addr_if))
{
    apu_.UseSound = false;
    bus_.LoadCartridge(path);
    ppu_.UseCGB = bus_.UseCGB;
    reset();
}

void Gameboy::ExecuteCommand(Command command) {
    switch (command) {
        case Command::Reset: {
            reset();
            break;
        }
        case Command::Frame: {
            frame();
            break;
        }
        case Command::Second: {
            for (int i = 0; i < 60; i++)
                frame();
            break;
        }
        case Command::Screenshot: {
            ExecuteCommand(Command::Second);
            screenshot();
            break;
        }
        case Command::Start: {
			action(3)
            break;
        }
        case Command::Select: {
            action(2);
            break;
        }
        case Command::B: {
            action(1);
            break;
        }
        case Command::A: {
            action(0);
            break;
        }
        case Command::Down: {
            direction(3);
            break;
        }
        case Command::Up: {
            direction(2);
            break;
        }
        case Command::Left: {
            direction(1);
            break;
        }
        case Command::Right: {
            direction(0);
            break;
        }
        case Command::Save: {
            save();
            bus_.battery_save();
            break;
        }
        case Command::Load: {
            load();
            break;
        }
        case Command::ReadSingle: {
            uint8_t data = 0xff;
            {
                std::stringstream ss;
                ss << std::hex << value_;
                uint16_t addr = 0;
                ss >> addr;
                data = bus_.Read(addr);
            }
            std::stringstream ss;
            ss << std::hex << (int)data;
            res_ = ss.str();
            std::cout << "reading   data: " << res_ << std::endl;
            break;
        }
        case Command::ReadString: {
            uint16_t addr = 0;
            {
                std::stringstream ss;
                ss << std::hex << value_;
                ss >> addr;
            }
            std::stringstream ss;
            for (int i = 0; i < 0xB; i++) {
                ss << bus_.Read(addr + i);
            }
            res_ = ss.str();
            std::cout << "reading from: " << addr << std::endl;
            break; 
        }
        default: {
            // ignore
            break;
        }
    }
}

void Gameboy::reset() {
    bus_.SoftReset();
    cpu_.Reset(true);
    timer_.Reset();
    ppu_.Reset();
}

void Gameboy::frame() {
    while (frame_clk_ < 70224)
        update();
    frame_clk_ = 0;
}

void Gameboy::update() {
    uint8_t old_if = interrupt_flag_;
    int clk = 0;
    if (!cpu_.skip_next_)
        clk = cpu_.Update();
    cpu_.skip_next_ = false;
    if (timer_.Update(clk, old_if)) {
        if (cpu_.halt_) {
            cpu_.halt_ = false;
            cpu_.skip_next_ = true;
        }
    }
    ppu_.Update(clk);
    apu_.Update(clk);
    frame_clk_ += clk;
}

void callback(void* context, void* data, int size) {
    std::string* str = (std::string*)context;
    std::stringstream ss;
    ss << std::hex;
    for (int i = 0; i < size; i++) {
        ss << std::setw(2) << std::setfill('0') << (int)(((uint8_t*)data)[i]);
    }
    *str =  ss.str();
}

void Gameboy::screenshot() {
    uint8_t* data = ppu_.GetScreenData();
    auto img_s = to_image_small(data);
    auto img_b = scale(img_s);
    auto data_b = to_bytes(img_b);
    stbi_write_png_to_func(&callback, &res_, 320, 288, 4, data_b.data(), 0);
}

void Gameboy::save() {
    std::ofstream ofs("save.sav", std::ios::binary);
    write16(cpu_.PC)
    write8(cpu_.A)
    write8(cpu_.F)
    write8(cpu_.B)
    write8(cpu_.C)
    write8(cpu_.D)
    write8(cpu_.E)
    write16(cpu_.SP)
    write8(cpu_.ime_)
    write8(cpu_.IE)
    write8(cpu_.halt_)
    ofs.write((char*)&bus_.hram_[0], 0x100);
    for (int i = 0; i < bus_.ram_banks_.size(); i++)
        ofs.write((char*)&bus_.ram_banks_[i][0], 0x2000);
    for (int i = 0; i < bus_.vram_banks_.size(); i++)
        ofs.write((char*)&bus_.vram_banks_[i][0], 0x2000);
    for (int i = 0; i < bus_.wram_banks_.size(); i++)
        ofs.write((char*)&bus_.wram_banks_[i][0], 0x1000);
    for (const auto& pal : bus_.BGPalettes) {
        ofs.write((char*)&pal[0], 2 * 4);
    }
    for (const auto& pal : bus_.OBJPalettes) {
        ofs.write((char*)&pal[0], 2 * 4);
    }
    ofs.write((char*)&bus_.oam_[0], bus_.oam_.size());
}

void Gameboy::load() {
    std::ifstream ifs("save.sav", std::ios::binary);
    ifs.read((char*)&cpu_.PC, 2);
    ifs.read((char*)&cpu_.A, 1);
    ifs.read((char*)&cpu_.F, 1);
    ifs.read((char*)&cpu_.B, 1);
    ifs.read((char*)&cpu_.C, 1);
    ifs.read((char*)&cpu_.D, 1);
    ifs.read((char*)&cpu_.E, 1);
    ifs.read((char*)&cpu_.SP, 2);
    ifs.read((char*)&cpu_.ime_, 1);
    ifs.read((char*)&cpu_.IE, 1);
    ifs.read((char*)&cpu_.halt_, 1);
    ifs.read((char*)&bus_.hram_[0], 0x100);
    for (int i = 0; i < bus_.ram_banks_.size(); i++)
        ifs.read((char*)&bus_.ram_banks_[i][0], 0x2000);
    for (int i = 0; i < bus_.vram_banks_.size(); i++)
        ifs.read((char*)&bus_.vram_banks_[i][0], 0x2000);
    for (int i = 0; i < bus_.wram_banks_.size(); i++)
        ifs.read((char*)&bus_.wram_banks_[i][0], 0x1000);
    for (const auto& pal : bus_.BGPalettes)
        ifs.read((char*)&pal[0], 2 * 4);
    for (const auto& pal : bus_.OBJPalettes)
        ifs.read((char*)&pal[0], 2 * 4);
    ifs.read((char*)&bus_.oam_[0], bus_.oam_.size());
}