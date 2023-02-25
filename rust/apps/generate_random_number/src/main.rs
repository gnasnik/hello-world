use rand::prelude::*;

fn main() {
    let mut rng = rand::thread_rng();

    let n1:u8 = rng.gen();
    let n2:i16 = rng.gen();

    println!("Random u8: {}", n1);
    println!("Random i16: {}", n2);
    println!("Random i32: {}", rng.gen::<i32>());
    println!("Random u32: {}", rng.gen::<u32>());
    println!("Random f64: {}", rng.gen::<f64>());
}
