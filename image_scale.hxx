#ifndef IMAGE_SCALE_HXX
#define IMAGE_SCALE_HXX
#include <cstdint>
#include <array>
#include <vector>
#include <assert.h>
struct P {
    uint8_t R, G, B, A;
};
using image_small = std::array<std::array<P, 160>, 144>;
using image_big = std::array<std::array<P, 320>, 288>;

image_small to_image_small(const uint8_t* data) {
    int size = 160 * 144 * 4;
    image_small ret;
    int x = 0;
    int y = 0;
    for (int i = 0; i < size; i += 4) {
        P cur_pixel = *reinterpret_cast<const P*>(&data[i]);
        ret[y][x] = cur_pixel;
        x++;
        if (x == 160) {
            x = 0;
            y++;
        }
    }
    assert(y == 144 && x == 0);
    return ret;
}

image_big scale(const image_small& image) {
    image_big ret;
    int x = 0;
    int y = 0;
    for (int i = 0; i < 144; i++) {
        for (int j = 0; j < 160; j++) {
            P cur_pixel = image[i][j];
            ret[y][x] = cur_pixel;
            ret[y][x + 1] = cur_pixel;
            ret[y + 1][x] = cur_pixel;
            ret[y + 1][x + 1] = cur_pixel;
            x += 2;
            if (x == 320) {
                x = 0;
                y += 2;
            }
        }
    }
    assert(y == 288 && x == 0);
    return ret;
}

void to_bytes(const image_big& image, uint8_t* ret) {
    for (int y = 0; y < 288; y++) {
        for (int x = 0; x < 320; x++) {
            ret[((y * 320) + x) * 4] = image[y][x].R;
            ret[((y * 320) + x) * 4 + 1] = image[y][x].G;
            ret[((y * 320) + x) * 4 + 2] = image[y][x].B;
            ret[((y * 320) + x) * 4 + 3] = image[y][x].A;
        }
    }
}

#endif