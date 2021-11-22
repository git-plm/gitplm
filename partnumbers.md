# Part number guidelines

Basic format: CCC-NNN-VVVV

- CCC: major category (RES, CAP, DIO, etc)
- NNN: incrementing sequential number for each part
- VVVV: variation to code variations of a parts typically with the **same
  datasheet** (resistance, capacitance, regulator voltage, IC package, etc.)

Each group of CCC parts is placed in its own schematic symbol library with the
same name.

The following CCC groups are suggested for electrical parts:

- ANA: op-amps, comparators, A/D, D/A
- CAP: capacitors
- CON: connectors
- DIO: diodes
- IND: inductors
- ICS: integrated circuits, mcus, etc
- OSC: oscillators
- PWR: relays, etc
- REG: regulators
- RES: resistors
- SWI: switch
- TRA: transistors, FETs
- TXT: test points/pads
- XTL: crystals

The following CCC groups are suggested for high level assemblies:

- PCB: Printed Circuit board
- ASY: Assembly (can be top level or subassembly -- typically represented by BOM
  and documentation)

Some additional guidelines:

- Every part number has the same number of characters in it (3-3-4). This makes
  sorting simpler and less chance of error.
- Character set is restricted to capital letters, digits, and hyphen.
- Avoid punctuation characters such as %, !, (, ., etc.

Examples:

With resistors, capacitors, and connectors, we encode the value and pin count in
the variation:

- 1K 0805 1%: RES-002-01K0
- 3.3K 0805 1%: RES-002-03K3
- 2.2K 0603 5%: RES-003-02K2 (note we bumped NNN to 003, because different
  package size)
- 10.3K high power 0603: RES-004-10K3 (different vendor/datasheet than RES-002,
  so we bump NNN)
- 2x10, 0.1 in header: CON-000-0020
- 2x12, 0.1 in header: CON-000-0024
- 1x10, 0.1 in header: CON-001-0010
- 1x20, 0.1 in header: CON-001-0020

With most ICs we simply enumerate all the variations of a particular IC in a
sequentially incrementing variation (we don't try to encode information)

- LM78xx SOT223 5V: REG-089-0000
- LM78xx DIP 5V: REG-089-0001
- LM78xx SOT223 3.3V: REG-089-0002
- LM78xx DIP 3.3V: REG-089-0003
- 3.3v switching reg, SSOP8: REG-002-0000
- 3.3v switching reg, S08: REG-002-0001
- STM32H7 in 44 pin package, 1M flash: MCU-001-0000
- STM32H7 in 44 pin package, 2M flash: MCU-001-0001
- STM32H7 in 208 pin package, 1M flash: MCU-001-0002
- STM32H7 in 208 pin package, 2M flash: MCU-001-0002
- STM32F3 in 44 pin package: MCU-002-0044 (not different base part, so bump NNN)

Many parts will not have any variations:

- 2N4401 DIODE: DIO-000-0000 (no variation information, that is fine)
- 2N2222 transistor: TRA-000-0000 (again, no variation info)

The variation section is only used in cases where a part with a single datasheet
has multiple variations. Variations are generally used to encode one parameter
with the most different variations -- for instance resistance with resistors. A
single datasheet may include 0603, 0805, and 1206 but for resistors, but take
out separate NNN part numbers for different package sizes because with
resistors, it makes the most sense to encode the resistance in the variation
(because there are lots of resistance values), not the package size. There are
relatively few package sizes for resistors so it makes sense to take out new NNN
numbers for different packages. However, for voltage regulators, it may make
sense to encode both the regulated voltage and the package in the variation,
because there is a relatively small number of combinations.

Generally we don't need to create house part numbers for every part variation --
only the ones we use. Resistors/caps may be an exception where we simply create
the entire series in the partmaster because it is easiest to just do once.

Resistor variations are encoded using the industry standard resistor variations.
R, K, and M are used to designate decimal place. Examples:

- 10ohm: 010R
- 10K: 010K
- 11.3K: 11K3
- 1.21M: 1M21

Capacitor variations are specified using 4 numbers: 3 significant digits+number
of zeros. These numbers are multiplied to get pico fareds.

The 4th digit signifies the multiplying factor, and letter R is decimal point.

Examples:

- 1002 = `100 x 10^2 = 10,000 pF = 10nF = 0.01uF`
- 1003 = 0.1uF

To figure out the extension, you can divide the capitance by 1pF to get the
number of pF. From this, you can visually tell what the variation should be.
Example:

`0.022uF = 0.022e-6/1e-12 = 22000`

So the extension would be `2202`

To work backwards, we would have `100 * 220 = 22,000pF/1e-6 = 0.022uF`.
