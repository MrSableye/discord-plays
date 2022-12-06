#include "gb_headlesswrapper.hxx"
#define STB_IMAGE_WRITE_IMPLEMENTATION
#include "stb_image_write.h"
#include "json.hpp"
#include <fstream>
#include <vector>
#include <iostream>
#include <image_scale.hxx>
using namespace nlohmann;
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

#define direction(index) { bus_.DirectionKeys &= (~(1UL << index)); \
            interrupt_flag_ |= IFInterrupt::JOYPAD; \
            for (int i = 0; i < 5; i++) \
                ExecuteCommand(Command::Frame); \
            bus_.DirectionKeys |= (1UL << index); \
            ExecuteCommand(Command::Second); }

#define action(index) { bus_.ActionKeys &= (~(1UL << index)); \
            interrupt_flag_ |= IFInterrupt::JOYPAD; \
            for (int i = 0; i < 5; i++) \
                ExecuteCommand(Command::Frame); \
            bus_.ActionKeys |= (1UL << index); \
            ExecuteCommand(Command::Second); }

struct Pokemon {
    uint8_t type;
    uint8_t item_held;
    uint8_t moves[4];
    uint8_t id[2];
    uint8_t exp[3];
    uint8_t hp_ev[2];
    uint8_t atk_ev[2];
    uint8_t def_ev[2];
    uint8_t spd_ev[2];
    uint8_t spc_ev[2];
    uint8_t atk_def_iv;
    uint8_t spd_spc_iv;
    uint8_t pp_moves[4];
    uint8_t happiness;
    uint8_t pokerus;
    uint8_t caught_data[2];
    uint8_t level;
    uint8_t status[2];
    uint8_t hp[2];
    uint8_t max_hp[2];
    uint8_t atk[2];
    uint8_t def[2];
    uint8_t spd[2];
    uint8_t spc_def[2];
    uint8_t spc_atk[2];
};
struct PokemonName {
    uint8_t name[0xB];
};
struct Pokeball {
    uint8_t type;
    uint8_t count;
};

Gameboy::Gameboy(std::string path) :
    channel_array_ptr_(std::make_shared<ChannelArray>()),
    bus_(channel_array_ptr_),
    apu_(channel_array_ptr_, bus_.GetReference(addr_NR52)),
    ppu_(bus_, nullptr),
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
            ppu_.Draw = true;
            frame();
            frame();
            screenshot();
            ppu_.Draw = false;
            break;
        }
        case Command::Start: {
            int count = std::atoi(value_.c_str());
            if ((count <= 0) || (count > 10))
                count = 1;
            for (int i = 0; i < count; i++)
			    action(3)
            break;
        }
        case Command::Select: {
            int count = std::atoi(value_.c_str());
            if ((count <= 0) || (count > 10))
                count = 1;
            for (int i = 0; i < count; i++)
                action(2);
            break;
        }
        case Command::B: {
            int count = std::atoi(value_.c_str());
            if ((count <= 0) || (count > 10))
                count = 1;
            for (int i = 0; i < count; i++)
                action(1);
            break;
        }
        case Command::A: {
            int count = std::atoi(value_.c_str());
            if ((count <= 0) || (count > 10))
                count = 1;
            for (int i = 0; i < count; i++)
                action(0);
            break;
        }
        case Command::Down: {
            int count = std::atoi(value_.c_str());
            if ((count <= 0) || (count > 10))
                count = 1;
            for (int i = 0; i < count; i++)
                direction(3);
            break;
        }
        case Command::Up: {
            int count = std::atoi(value_.c_str());
            if ((count <= 0) || (count > 10))
                count = 1;
            for (int i = 0; i < count; i++)
                direction(2);
            break;
        }
        case Command::Left: {
            int count = std::atoi(value_.c_str());
            if ((count <= 0) || (count > 10))
                count = 1;
            for (int i = 0; i < count; i++)
                direction(1);
            break;
        }
        case Command::Right: {
            int count = std::atoi(value_.c_str());
            if ((count <= 0) || (count > 10))
                count = 1;
            for (int i = 0; i < count; i++)
                direction(0);
            break;
        }
        case Command::Save: {
            save();
            bus_.battery_save();
            break;
        }
        case Command::StartSave: {
            value_ = "";
            ExecuteCommand(Command::Start);
            // Change all options to "Save"
            bus_.Write(0xCF2A, 0x04);
            bus_.Write(0xCF2B, 0x04);
            bus_.Write(0xCF2C, 0x04);
            bus_.Write(0xCF2D, 0x04);
            bus_.Write(0xCF2E, 0x04);
            bus_.Write(0xCF2F, 0x04);
            bus_.Write(0xCF30, 0x04);
            bus_.Write(0xCF31, 0x04);
            ExecuteCommand(Command::A);
            ExecuteCommand(Command::A);
            ExecuteCommand(Command::A);
            ExecuteCommand(Command::A);
            ExecuteCommand(Command::A);
            ExecuteCommand(Command::Frame);
            bus_.battery_save();
            break;
        }
        case Command::Load: {
            load();
            break;
        }
        case Command::GetParty: {
            get_party();
            break;
        }
        case Command::GetBalls: {
            get_balls();
            break;
        }
        case Command::GetTrainer: {
            get_trainer();
            break;
        }
        case Command::GetMap: {
            value_ = "";
            ExecuteCommand(Command::Start);
            // Change all options to "Pokegear"
            bus_.Write(0xCF2A, 0x07);
            bus_.Write(0xCF2B, 0x07);
            bus_.Write(0xCF2C, 0x07);
            bus_.Write(0xCF2D, 0x07);
            bus_.Write(0xCF2E, 0x07);
            bus_.Write(0xCF2F, 0x07);
            bus_.Write(0xCF30, 0x07);
            bus_.Write(0xCF31, 0x07);
            ExecuteCommand(Command::A);
            ExecuteCommand(Command::Right);
            ExecuteCommand(Command::Frame);
            ExecuteCommand(Command::Screenshot);
            ExecuteCommand(Command::B);
            ExecuteCommand(Command::B);
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

void Gameboy::get_party() {
    int poke_size = bus_.Read(0xDA22);
    std::vector<json> pokemen;
    for (int i = 0; i < poke_size; i++) {
        uint16_t addr = 0xDA2A + (sizeof(Pokemon) * i);
        Pokemon poke;
        uint8_t* mem = &bus_.fast_redirect_address(addr);
        memcpy(&poke, mem, sizeof(Pokemon));
        json o;
        o["Type"] = poke_type_to_name(poke.type);
        o["Status"] = static_cast<uint16_t>(poke.status[0] << 8) | poke.status[1];
        o["Level"] = poke.level;
        o["Exp"] = 
            poke.exp[0] << 16 
            | poke.exp[1] << 8
            | poke.exp[2];
        o["Hp"] = poke.hp[0] << 8 | poke.hp[1];
        o["MaxHp"] = poke.max_hp[0] << 8 | poke.max_hp[1];
        addr = 0xDB8C + i * 0xB;
        PokemonName pname;
        mem = &bus_.fast_redirect_address(addr);
        memcpy(&pname.name, mem, sizeof(PokemonName));
        o["Name"] = poke_to_ascii(pname.name, 0xB);
        pokemen.push_back(o);
    }
    json arr = pokemen;
    res_ = arr.dump();
}

void Gameboy::get_balls() {
    std::vector<Pokeball> pokeballs;
    int ball_size = bus_.Read(0xD5FC);
    int sum = 0;
    std::vector<json> balls;
    balls.resize(ball_size);
    for (int i = 0; i < ball_size; i++) {
        uint16_t addr = 0xD5FD + (sizeof(Pokeball) * i);
        Pokeball ball;
        uint8_t* mem = &bus_.fast_redirect_address(addr);
        memcpy(&ball, mem, sizeof(Pokeball));
        sum += ball.count;
        json b;
        b["Name"] = poke_item_to_name(ball.type);
        b["Count"] = ball.count;
        balls[i] = b;
    }
    json arr;
    arr["Count"] = sum;
    arr["Balls"] = balls;
    res_ = arr.dump();
}

void Gameboy::get_trainer() {
    json t;
    uint8_t ch[7], chr[7];
    uint8_t* mem = &bus_.fast_redirect_address(0xD1A3);
    uint8_t* memr = &bus_.fast_redirect_address(0xD1B9);
    memcpy(ch, mem, 7);
    memcpy(chr, memr, 7);
    uint8_t m[4];
    uint8_t* memm = &bus_.fast_redirect_address(0xD573);
    memcpy(m, memm, 3);
    uint32_t money = m[0] << 16 | m[1] << 8 | m[2];
    
    t["Name"] = poke_to_ascii(ch, 7);
    t["Rival"] = poke_to_ascii(chr, 7);
    t["Money"] = money;
    t["Johto"] = bus_.Read(0xD57C);
    t["Kanto"] = bus_.Read(0xD57D);
    res_ = t.dump();
}

std::string Gameboy::poke_item_to_name(uint8_t type) {
    switch (type) {
        case 0x00: return "?";
        case 0x01: return "Master Ball";
        case 0x02: return "Ultra Ball";
        case 0x03: return "BrightPowder";
        case 0x04: return "Great Ball";
        case 0x05: return "Poké Ball";
        case 0x06: return "Teru-sama";
        case 0x07: return "Bicycle";
        case 0x08: return "Moon Stone";
        case 0x09: return "Antidote";
        case 0x0A: return "Burn Heal";
        case 0x0B: return "Ice Heal";
        case 0x0C: return "Awakening";
        case 0x0D: return "Parlyz Heal";
        case 0x0E: return "Full Restore";
        case 0x0F: return "Max Potion";
        case 0x10: return "Hyper Potion";
        case 0x11: return "Super Potion";
        case 0x12: return "Potion";
        case 0x13: return "Escape Rope";
        case 0x14: return "Repel";
        case 0x15: return "Max Elixer";
        case 0x16: return "Fire Stone";
        case 0x17: return "Thunderstone";
        case 0x18: return "Water Stone";
        case 0x19: return "Teru-sama";
        case 0x1A: return "HP Up";
        case 0x1B: return "Protein";
        case 0x1C: return "Iron";
        case 0x1D: return "Carbos";
        case 0x1E: return "Lucky Punch";
        case 0x1F: return "Calcium";
        case 0x20: return "Rare Candy";
        case 0x21: return "X Accuracy";
        case 0x22: return "Leaf Stone";
        case 0x23: return "Metal Powder";
        case 0x24: return "Nugget";
        case 0x25: return "Poké Doll";
        case 0x26: return "Full Heal";
        case 0x27: return "Revive";
        case 0x28: return "Max Revive";
        case 0x29: return "Guard Spec.";
        case 0x2A: return "Super Repel";
        case 0x2B: return "Max Repel";
        case 0x2C: return "Dire Hit";
        case 0x2D: return "Teru-sama";
        case 0x2E: return "Fresh Water";
        case 0x2F: return "Soda Pop";
        case 0x30: return "Lemonade";
        case 0x31: return "X Attack";
        case 0x32: return "Teru-sama";
        case 0x33: return "X Defend";
        case 0x34: return "X Speed";
        case 0x35: return "X Special";
        case 0x36: return "Coin Case";
        case 0x37: return "Itemfinder";
        case 0x38: return "Teru-sama";
        case 0x39: return "Exp.Share";
        case 0x3A: return "Old Rod";
        case 0x3B: return "Good Rod";
        case 0x3C: return "Silver Leaf";
        case 0x3D: return "Super Rod";
        case 0x3E: return "PP Up";
        case 0x3F: return "Ether";
        case 0x40: return "Max Ether";
        case 0x41: return "Elixer";
        case 0x42: return "Red Scale";
        case 0x43: return "SecretPotion";
        case 0x44: return "S.S. Ticket";
        case 0x45: return "Mystery Egg";
        case 0x46: return "Clear Bell*";
        case 0x47: return "Silver Wing";
        case 0x48: return "Moomoo Milk";
        case 0x49: return "Quick Claw";
        case 0x4A: return "PSNCureBerry";
        case 0x4B: return "Gold Leaf";
        case 0x4C: return "Soft Sand";
        case 0x4D: return "Sharp Beak";
        case 0x4E: return "PRZCureBerry";
        case 0x4F: return "Burnt Berry";
        case 0x50: return "Ice Berry";
        case 0x51: return "Poison Barb";
        case 0x52: return "King's Rock";
        case 0x53: return "Bitter Berry";
        case 0x54: return "Mint Berry";
        case 0x55: return "Red Apricorn";
        case 0x56: return "TinyMushroom";
        case 0x57: return "Big Mushroom";
        case 0x58: return "SilverPowder";
        case 0x59: return "Blu Apricorn";
        case 0x5A: return "Teru-sama";
        case 0x5B: return "Amulet Coin";
        case 0x5C: return "Ylw Apricorn";
        case 0x5D: return "Grn Apricorn";
        case 0x5E: return "Cleanse Tag";
        case 0x5F: return "Mystic Water";
        case 0x60: return "TwistedSpoon";
        case 0x61: return "Wht Apricorn";
        case 0x62: return "Blackbelt";
        case 0x63: return "Blk Apricorn";
        case 0x64: return "Teru-sama";
        case 0x65: return "Pnk Apricorn";
        case 0x66: return "BlackGlasses";
        case 0x67: return "SlowpokeTail";
        case 0x68: return "Pink Bow";
        case 0x69: return "Stick";
        case 0x6A: return "Smoke Ball";
        case 0x6B: return "NeverMeltIce";
        case 0x6C: return "Magnet";
        case 0x6D: return "MiracleBerry";
        case 0x6E: return "Pearl";
        case 0x6F: return "Big Pearl";
        case 0x70: return "Everstone";
        case 0x71: return "Spell Tag";
        case 0x72: return "RageCandyBar";
        case 0x73: return "GS Ball*";
        case 0x74: return "Blue Card*";
        case 0x75: return "Miracle Seed";
        case 0x76: return "Thick Club";
        case 0x77: return "Focus Band";
        case 0x78: return "Teru-sama";
        case 0x79: return "EnergyPowder";
        case 0x7A: return "Energy Root";
        case 0x7B: return "Heal Powder";
        case 0x7C: return "Revival Herb";
        case 0x7D: return "Hard Stone";
        case 0x7E: return "Lucky Egg";
        case 0x7F: return "Card Key";
        case 0x80: return "Machine Part";
        case 0x81: return "Egg Ticket*";
        case 0x82: return "Lost Item";
        case 0x83: return "Stardust";
        case 0x84: return "Star Piece";
        case 0x85: return "Basement Key";
        case 0x86: return "Pass";
        case 0x87: return "Teru-sama";
        case 0x88: return "Teru-sama";
        case 0x89: return "Teru-sama";
        case 0x8A: return "Charcoal";
        case 0x8B: return "Berry Juice";
        case 0x8C: return "Scope Lens";
        case 0x8D: return "Teru-sama";
        case 0x8E: return "Teru-sama";
        case 0x8F: return "Metal Coat";
        case 0x90: return "Dragon Fang";
        case 0x91: return "Teru-sama";
        case 0x92: return "Leftovers";
        case 0x93: return "Teru-sama";
        case 0x94: return "Teru-sama";
        case 0x95: return "Teru-sama";
        case 0x96: return "MysteryBerry";
        case 0x97: return "Dragon Scale";
        case 0x98: return "Berserk Gene";
        case 0x99: return "Teru-sama";
        case 0x9A: return "Teru-sama";
        case 0x9B: return "Teru-sama";
        case 0x9C: return "Sacred Ash";
        case 0x9D: return "Heavy Ball";
        case 0x9E: return "Flower Mail";
        case 0x9F: return "Level Ball";
        case 0xA0: return "Lure Ball";
        case 0xA1: return "Fast Ball";
        case 0xA2: return "Teru-sama";
        case 0xA3: return "Light Ball";
        case 0xA4: return "Friend Ball";
        case 0xA5: return "Moon Ball";
        case 0xA6: return "Love Ball";
        case 0xA7: return "Normal Box";
        case 0xA8: return "Gorgeous Box";
        case 0xA9: return "Sun Stone";
        case 0xAA: return "Polkadot Bow";
        case 0xAB: return "Teru-sama";
        case 0xAC: return "Up-Grade";
        case 0xAD: return "Berry";
        case 0xAE: return "Gold Berry";
        case 0xAF: return "SquirtBottle";
        case 0xB0: return "Teru-sama";
        case 0xB1: return "Park Ball";
        case 0xB2: return "Rainbow Wing";
        case 0xB3: return "Teru-sama";
        case 0xB4: return "Brick Piece";
        case 0xB5: return "Surf Mail";
        case 0xB6: return "Litebluemail";
        case 0xB7: return "Portraitmail";
        case 0xB8: return "Lovely Mail";
        case 0xB9: return "Eon Mail";
        case 0xBA: return "Morph Mail";
        case 0xBB: return "Bluesky Mail";
        case 0xBC: return "Music Mail";
        case 0xBD: return "Mirage Mail";
        case 0xBE: return "Teru-sama";
        case 0xBF: return "TM01";
        case 0xC0: return "TM02";
        case 0xC1: return "TM03";
        case 0xC2: return "TM04";
        case 0xC3: return "TM04";
        case 0xC4: return "TM05";
        case 0xC5: return "TM06";
        case 0xC6: return "TM07";
        case 0xC7: return "TM08";
        case 0xC8: return "TM09";
        case 0xC9: return "TM10";
        case 0xCA: return "TM11";
        case 0xCB: return "TM12";
        case 0xCC: return "TM13";
        case 0xCD: return "TM14";
        case 0xCE: return "TM15";
        case 0xCF: return "TM16";
        case 0xD0: return "TM17";
        case 0xD1: return "TM18";
        case 0xD2: return "TM19";
        case 0xD3: return "TM20";
        case 0xD4: return "TM21";
        case 0xD5: return "TM22";
        case 0xD6: return "TM23";
        case 0xD7: return "TM24";
        case 0xD8: return "TM25";
        case 0xD9: return "TM26";
        case 0xDA: return "TM27";
        case 0xDB: return "TM28";
        case 0xDC: return "TM28";
        case 0xDD: return "TM29";
        case 0xDE: return "TM30";
        case 0xDF: return "TM31";
        case 0xE0: return "TM32";
        case 0xE1: return "TM33";
        case 0xE2: return "TM34";
        case 0xE3: return "TM35";
        case 0xE4: return "TM36";
        case 0xE5: return "TM37";
        case 0xE6: return "TM38";
        case 0xE7: return "TM39";
        case 0xE8: return "TM40";
        case 0xE9: return "TM41";
        case 0xEA: return "TM42";
        case 0xEB: return "TM43";
        case 0xEC: return "TM44";
        case 0xED: return "TM45";
        case 0xEE: return "TM46";
        case 0xEF: return "TM47";
        case 0xF0: return "TM48";
        case 0xF1: return "TM49";
        case 0xF2: return "TM50";
        case 0xF3: return "HM01";
        case 0xF4: return "HM02";
        case 0xF5: return "HM03";
        case 0xF6: return "HM04";
        case 0xF7: return "HM05";
        case 0xF8: return "HM06";
        case 0xF9: return "HM07";
        case 0xFA: return "HM08";
        case 0xFB: return "HM09";
        case 0xFC: return "HM10";
        case 0xFD: return "HM11";
        case 0xFE: return "HM12";
        case 0xFF: return "Cancel";
    }
    return "???";
}

std::string Gameboy::poke_to_ascii(uint8_t* data, int size) {
    std::string ret;
    for (int i = 0; i < size; i++) {
        auto ch = data[i];
        switch (ch) {
            case 0x4F: ret += "="; break;
            case 0x57: ret += "#"; break;
            case 0x51: ret += "*"; break;
            case 0x52: ret += "A1"; break;
            case 0x53: ret += "A2"; break;
            case 0x54: ret += "PK"; break;
            case 0x55: ret += "+"; break;
            case 0x58: ret += "$"; break;
            case 0x7F: ret += " "; break;
            case 0x80: ret += "A"; break;
            case 0x81: ret += "B"; break;
            case 0x82: ret += "C"; break;
            case 0x83: ret += "D"; break;
            case 0x84: ret += "E"; break;
            case 0x85: ret += "F"; break;
            case 0x86: ret += "G"; break;
            case 0x87: ret += "H"; break;
            case 0x88: ret += "I"; break;
            case 0x89: ret += "J"; break;
            case 0x8A: ret += "K"; break;
            case 0x8B: ret += "L"; break;
            case 0x8C: ret += "M"; break;
            case 0x8D: ret += "N"; break;
            case 0x8E: ret += "O"; break;
            case 0x8F: ret += "P"; break;
            case 0x90: ret += "Q"; break;
            case 0x91: ret += "R"; break;
            case 0x92: ret += "S"; break;
            case 0x93: ret += "T"; break;
            case 0x94: ret += "U"; break;
            case 0x95: ret += "V"; break;
            case 0x96: ret += "W"; break;
            case 0x97: ret += "X"; break;
            case 0x98: ret += "Y"; break;
            case 0x99: ret += "Z"; break;
            case 0x9C: ret += ":"; break;
            case 0xA0: ret += "a"; break;
            case 0xA1: ret += "b"; break;
            case 0xA2: ret += "c"; break;
            case 0xA3: ret += "d"; break;
            case 0xA4: ret += "e"; break;
            case 0xA5: ret += "f"; break;
            case 0xA6: ret += "g"; break;
            case 0xA7: ret += "h"; break;
            case 0xA8: ret += "i"; break;
            case 0xA9: ret += "j"; break;
            case 0xAA: ret += "k"; break;
            case 0xAB: ret += "l"; break;
            case 0xAC: ret += "m"; break;
            case 0xAD: ret += "n"; break;
            case 0xAE: ret += "o"; break;
            case 0xAF: ret += "p"; break;
            case 0xB0: ret += "q"; break;
            case 0xB1: ret += "r"; break;
            case 0xB2: ret += "s"; break;
            case 0xB3: ret += "t"; break;
            case 0xB4: ret += "u"; break;
            case 0xB5: ret += "v"; break;
            case 0xB6: ret += "w"; break;
            case 0xB7: ret += "x"; break;
            case 0xB8: ret += "y"; break;
            case 0xB9: ret += "z"; break;
            case 0xBA: ret += ","; break;
            case 0xBC: ret += "'l"; break;
            case 0xBD: ret += "'s"; break;
            case 0xBE: ret += "'t"; break;
            case 0xBF: ret += "'v"; break;
            case 0xE0: ret += "'"; break;
            case 0xE1: ret += "PK"; break;
            case 0xE2: ret += "MN"; break;
            case 0xE3: ret += "-"; break;
            case 0xE4: ret += "'r"; break;
            case 0xE5: ret += "'m"; break;
            case 0xE6: ret += "?"; break;
            case 0xE7: ret += "!"; break;
            case 0xE8: ret += "."; break;
            case 0xF4: ret += ","; break;
            case 0xF6: ret += "0"; break;
            case 0xF7: ret += "1"; break;
            case 0xF8: ret += "2"; break;
            case 0xF9: ret += "3"; break;
            case 0xFA: ret += "4"; break;
            case 0xFB: ret += "5"; break;
            case 0xFC: ret += "6"; break;
            case 0xFD: ret += "7"; break;
            case 0xFE: ret += "8"; break;
            case 0xFF: ret += "9"; break;
        }
    }
    return ret;
}

std::string Gameboy::poke_type_to_name(uint8_t type) {
    switch (type) {
        case 0x3F: return "Abra";
        case 0x8E: return "Aerodactyl";
        case 0xBE: return "Aipom";
        case 0x41: return "Alakazam";
        case 0xB5: return "Ampharos";
        case 0x18: return "Arbok";
        case 0x3B: return "Arcanine";
        case 0xA8: return "Ariados";
        case 0x90: return "Articuno";
        case 0xB8: return "Azumarill";
        case 0x99: return "Bayleef";
        case 0x0F: return "Beedrill";
        case 0xB6: return "Bellossom";
        case 0x45: return "Bellsprout";
        case 0x09: return "Blastoise";
        case 0xF2: return "Blissey";
        case 0x01: return "Bulbasaur";
        case 0x0C: return "Butterfree";
        case 0x0A: return "Caterpie";
        case 0xFB: return "Celebi";
        case 0x71: return "Chansey";
        case 0x06: return "Charizard";
        case 0x04: return "Charmander";
        case 0x05: return "Charmeleon";
        case 0x98: return "Chikorita";
        case 0xAA: return "Chinchou";
        case 0x24: return "Clefable";
        case 0x23: return "Clefairy";
        case 0xAD: return "Cleffa";
        case 0x5B: return "Cloyster";
        case 0xDE: return "Corsola";
        case 0xA9: return "Crobat";
        case 0x9F: return "Croconaw";
        case 0x68: return "Cubone";
        case 0x9B: return "Cyndaquil";
        case 0xE1: return "Delibird";
        case 0x57: return "Dewgong";
        case 0x32: return "Diglett";
        case 0x84: return "Ditto";
        case 0x55: return "Dodrio";
        case 0x54: return "Doduo";
        case 0xE8: return "Donphan";
        case 0x94: return "Dragonair";
        case 0x95: return "Dragonite";
        case 0x93: return "Dratini";
        case 0x60: return "Drowzee";
        case 0x33: return "Dugtrio";
        case 0xCE: return "Dunsparce";
        case 0x85: return "Eevee";
        case 0x17: return "Ekans";
        case 0x7D: return "Electabuzz";
        case 0x65: return "Electrode";
        case 0xEF: return "Elekid";
        case 0xF4: return "Entei";
        case 0xC4: return "Espeon";
        case 0x66: return "Exeggcute";
        case 0x67: return "Exeggutor";
        case 0x53: return "Farfetch’d";
        case 0x16: return "Fearow";
        case 0xA0: return "Feraligatr";
        case 0xB4: return "Flaaffy";
        case 0x88: return "Flareon";
        case 0xCD: return "Forretress";
        case 0xA2: return "Furret";
        case 0x5C: return "Gastly";
        case 0x5E: return "Gengar";
        case 0x4A: return "Geodude";
        case 0xCB: return "Girafarig";
        case 0xCF: return "Gligar";
        case 0x2C: return "Gloom";
        case 0x2A: return "Golbat";
        case 0x76: return "Goldeen";
        case 0x37: return "Golduck";
        case 0x4C: return "Golem";
        case 0xD2: return "Granbull";
        case 0x4B: return "Graveler";
        case 0x58: return "Grimer";
        case 0x3A: return "Growlithe";
        case 0x82: return "Gyarados";
        case 0x5D: return "Haunter";
        case 0xD6: return "Heracross";
        case 0x6B: return "Hitmonchan";
        case 0x6A: return "Hitmonlee";
        case 0xED: return "Hitmontop";
        case 0xFA: return "Ho-oh";
        case 0xA3: return "Hoothoot";
        case 0xBB: return "Hoppip";
        case 0x74: return "Horsea";
        case 0xE5: return "Houndoom";
        case 0xE4: return "Houndour";
        case 0x61: return "Hypno";
        case 0xAE: return "Igglybuff";
        case 0x02: return "Ivysaur";
        case 0x27: return "Jigglypuff";
        case 0x87: return "Jolteon";
        case 0xBD: return "Jumpluff";
        case 0x7C: return "Jynx";
        case 0x8C: return "Kabuto";
        case 0x8D: return "Kabutops";
        case 0x40: return "Kadabra";
        case 0x0E: return "Kakuna";
        case 0x73: return "Kangaskhan";
        case 0xE6: return "Kingdra";
        case 0x63: return "Kingler";
        case 0x6D: return "Koffing";
        case 0x62: return "Krabby";
        case 0xAB: return "Lanturn";
        case 0x83: return "Lapras";
        case 0xF6: return "Larvitar";
        case 0xA6: return "Ledian";
        case 0xA5: return "Ledyba";
        case 0x6C: return "Lickitung";
        case 0xF9: return "Lugia";
        case 0x44: return "Machamp";
        case 0x43: return "Machoke";
        case 0x42: return "Machop";
        case 0xF0: return "Magby";
        case 0xDB: return "Magcargo";
        case 0x81: return "Magikarp";
        case 0x7E: return "Magmar";
        case 0x51: return "Magnemite";
        case 0x52: return "Magneton";
        case 0x38: return "Mankey";
        case 0xE2: return "Mantine";
        case 0xB3: return "Mareep";
        case 0xB7: return "Marill";
        case 0x69: return "Marowak";
        case 0x9A: return "Meganium";
        case 0x34: return "Meowth";
        case 0x0B: return "Metapod";
        case 0x97: return "Mew";
        case 0x96: return "Mewtwo";
        case 0xF1: return "Miltank";
        case 0xC8: return "Misdreavus";
        case 0x92: return "Moltres";
        case 0x7A: return "Mr. Mime";
        case 0x59: return "Muk";
        case 0xC6: return "Murkrow";
        case 0xB1: return "Natu";
        case 0x22: return "Nidoking";
        case 0x1F: return "Nidoqueen";
        case 0x1D: return "Nidoran F";
        case 0x20: return "Nidoran M";
        case 0x1E: return "Nidorina";
        case 0x21: return "Nidorino";
        case 0x26: return "Ninetales";
        case 0xA4: return "Noctowl";
        case 0xE0: return "Octillery";
        case 0x2B: return "Oddish";
        case 0x8A: return "Omanyte";
        case 0x8B: return "Omastar";
        case 0x5F: return "Onix";
        case 0x2E: return "Paras";
        case 0x2F: return "Parasect";
        case 0x35: return "Persian";
        case 0xE7: return "Phanpy";
        case 0xAC: return "Pichu";
        case 0x12: return "Pidgeot";
        case 0x11: return "Pidgeotto";
        case 0x10: return "Pidgey";
        case 0x19: return "Pikachu";
        case 0xDD: return "Piloswine";
        case 0xCC: return "Pineco";
        case 0x7F: return "Pinsir";
        case 0xBA: return "Politoed";
        case 0x3C: return "Poliwag";
        case 0x3D: return "Poliwhirl";
        case 0x3E: return "Poliwrath";
        case 0x4D: return "Ponyta";
        case 0x89: return "Porygon";
        case 0xE9: return "Porygon2";
        case 0x39: return "Primeape";
        case 0x36: return "Psyduck";
        case 0xF7: return "Pupitar";
        case 0xC3: return "Quagsire";
        case 0x9C: return "Quilava";
        case 0xD3: return "Quilfish";
        case 0x1A: return "Raichu";
        case 0xF3: return "Raikou";
        case 0x4E: return "Rapidash";
        case 0x14: return "Raticate";
        case 0x13: return "Rattata";
        case 0xDF: return "Remoraid";
        case 0x70: return "Rhydon";
        case 0x6F: return "Rhyhorn";
        case 0x1B: return "Sandshrew";
        case 0x1C: return "Sandslash";
        case 0xD4: return "Scizor";
        case 0x7B: return "Scyther";
        case 0x75: return "Seadra";
        case 0x77: return "Seaking";
        case 0x56: return "Seel";
        case 0xA1: return "Sentret";
        case 0x5A: return "Shellder";
        case 0xD5: return "Shuckle";
        case 0xE3: return "Skarmory";
        case 0xBC: return "Skiploom";
        case 0x50: return "Slowbro";
        case 0xC7: return "Slowking";
        case 0x4F: return "Slowpoke";
        case 0xDA: return "Slugma";
        case 0xEB: return "Smeargle";
        case 0xEE: return "Smoochum";
        case 0xD7: return "Sneasel";
        case 0x8F: return "Snorlax";
        case 0xD1: return "Snubull";
        case 0x15: return "Spearow";
        case 0xA7: return "Spinarak";
        case 0x07: return "Squirtle";
        case 0xEA: return "Stantler";
        case 0x79: return "Starmie";
        case 0x78: return "Staryu";
        case 0xD0: return "Steelix";
        case 0xB9: return "Sudwoodo";
        case 0xF5: return "Suicune";
        case 0xC0: return "Sunflora";
        case 0xBF: return "Sunkern";
        case 0xDC: return "Swinub";
        case 0x72: return "Tangela";
        case 0x80: return "Tauros";
        case 0xD8: return "Teddiursa";
        case 0x48: return "Tentacool";
        case 0x49: return "Tentacruel";
        case 0xAF: return "Togepi";
        case 0xB0: return "Togetic";
        case 0x9E: return "Totodile";
        case 0x9D: return "Typhlosion";
        case 0xF8: return "Tyranitar";
        case 0xEC: return "Tyrogue";
        case 0xC5: return "Umbreon";
        case 0xC9: return "Unown";
        case 0xD9: return "Ursaring";
        case 0x86: return "Vaporeon";
        case 0x31: return "Venomoth";
        case 0x30: return "Venonat";
        case 0x03: return "Venusaur";
        case 0x47: return "Victreebel";
        case 0x2D: return "Vileplume";
        case 0x64: return "Voltorb";
        case 0x25: return "Vulpix";
        case 0x08: return "Wartortle";
        case 0x0D: return "Weedle";
        case 0x46: return "Weepinbell";
        case 0x6E: return "Weezing";
        case 0x28: return "Wigglytuff";
        case 0xCA: return "Wobbuffet";
        case 0xC2: return "Wooper";
        case 0xB2: return "Xatu";
        case 0xC1: return "Yanma";
        case 0x91: return "Zapdos";
        case 0x29: return "Zubat";
    }
    return "???";
}